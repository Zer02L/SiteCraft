package importer

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"sitecraft/internal/data"
	"sitecraft/internal/media"
	"sitecraft/internal/models"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

var (
	reGutenbergComments = regexp.MustCompile(`<!--\s*\/?wp:[^\>]*-->`)
	reCaption           = regexp.MustCompile(`(?s)\[caption[^\]]*\](.*?)\[/caption\]`)
	reImgTag            = regexp.MustCompile(`(?i)<img[^>]+>`)
	reGallery           = regexp.MustCompile(`\[gallery[^\]]*\]`)
	reEmbed             = regexp.MustCompile(`(?s)\[embed[^\]]*\](.*?)\[/embed\]`)
	reAudioVideo        = regexp.MustCompile(`\[(audio|video)\s+src="([^"]+)"[^\]]*\](?:\[/(?:audio|video)\])?`)
	reRemainingTags     = regexp.MustCompile(`\[/?[a-zA-Z0-9_\-\:]+(?:\s+[^\]]*)?\]`)
	reExcessiveSpaces   = regexp.MustCompile(`\n{3,}`)
)

// Options configures the WordPress import process.
type Options struct {
	CreateRedirects  bool
	ImportPosts      bool
	ImportPages      bool
	CleanShortcodes  bool
	SanitizeHTML     bool
	DownloadMedia    bool
	DefaultAuthor    string
	PublishStatus    string // "all", "published_only", or empty for default
}

// DefaultOptions returns recommended options for migrating from WordPress.
func DefaultOptions() Options {
	return Options{
		CreateRedirects: true,
		ImportPosts:     true,
		ImportPages:     true,
		CleanShortcodes: true,
		SanitizeHTML:    true,
		DownloadMedia:   false,
		DefaultAuthor:   "Редакция",
		PublishStatus:   "all",
	}
}

// Progress tracks the live state of an active import job.
type Progress struct {
	TotalItems       int      `json:"total_items"`
	ProcessedItems   int      `json:"processed_items"`
	PostsImported    int      `json:"posts_imported"`
	PagesImported    int      `json:"pages_imported"`
	RedirectsCreated int      `json:"redirects_created"`
	MediaImported    int      `json:"media_imported"`
	Errors           []string `json:"errors"`
	CurrentItem      string   `json:"current_item"`
	IsComplete       bool     `json:"is_complete"`
	Percentage       int      `json:"percentage"`
}

// ProgressCallback is invoked as the import progresses.
type ProgressCallback func(p Progress)

// Importer orchestrates parsing WordPress WXR exports and persisting data into SiteCraft.
type Importer struct {
	repo      *data.Repository
	mediaSvc  *media.Service
	sanitizer *bluemonday.Policy
	client    *http.Client
}

// New creates an Importer instance.
func New(repo *data.Repository, mediaSvc *media.Service) *Importer {
	return &Importer{
		repo:      repo,
		mediaSvc:  mediaSvc,
		sanitizer: createSanitizer(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func createSanitizer() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowElements("figure", "figcaption", "article", "section", "header", "footer", "aside")
	p.AllowAttrs("class", "id", "style").Globally()
	p.AllowAttrs("src", "alt", "title", "width", "height", "loading").OnElements("img")
	p.AllowAttrs("href", "target", "rel", "title").OnElements("a")
	p.AllowAttrs("controls", "src", "preload").OnElements("audio", "video")
	p.AllowAttrs("colspan", "rowspan").OnElements("td", "th")
	return p
}

// WXRCategory represents category or post_tag in WordPress export.
type WXRCategory struct {
	Domain   string `xml:"domain,attr"`
	Nicename string `xml:"nicename,attr"`
	Name     string `xml:",chardata"`
}

// WXRItem holds all extracted fields from a WordPress `<item>`.
type WXRItem struct {
	Title         string
	Link          string
	PubDate       string
	Creator       string
	Description   string
	Content       string
	Excerpt       string
	PostID        string
	PostDate      string
	PostName      string
	PostType      string
	Status        string
	AttachmentURL string
	Categories    []WXRCategory
	PostMeta      map[string]string
}

// UnmarshalXML implements custom resilient parsing for `<item>` supporting all WXR variants.
func (item *WXRItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	item.PostMeta = make(map[string]string)
	var currentMetaKey, currentMetaVal string
	inPostMeta := false

	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch el := token.(type) {
		case xml.StartElement:
			local := el.Name.Local
			switch local {
			case "title":
				_ = d.DecodeElement(&item.Title, &el)
			case "link":
				_ = d.DecodeElement(&item.Link, &el)
			case "pubDate":
				_ = d.DecodeElement(&item.PubDate, &el)
			case "creator":
				_ = d.DecodeElement(&item.Creator, &el)
			case "description":
				_ = d.DecodeElement(&item.Description, &el)
			case "encoded":
				var val string
				_ = d.DecodeElement(&val, &el)
				if strings.Contains(el.Name.Space, "excerpt") {
					item.Excerpt = val
				} else {
					item.Content = val
				}
			case "post_id":
				_ = d.DecodeElement(&item.PostID, &el)
			case "post_date":
				_ = d.DecodeElement(&item.PostDate, &el)
			case "post_name":
				_ = d.DecodeElement(&item.PostName, &el)
			case "post_type":
				_ = d.DecodeElement(&item.PostType, &el)
			case "status":
				_ = d.DecodeElement(&item.Status, &el)
			case "attachment_url":
				_ = d.DecodeElement(&item.AttachmentURL, &el)
			case "category":
				var cat WXRCategory
				_ = d.DecodeElement(&cat, &el)
				item.Categories = append(item.Categories, cat)
			case "postmeta":
				inPostMeta = true
				currentMetaKey = ""
				currentMetaVal = ""
			case "meta_key":
				if inPostMeta {
					_ = d.DecodeElement(&currentMetaKey, &el)
				}
			case "meta_value":
				if inPostMeta {
					_ = d.DecodeElement(&currentMetaVal, &el)
				}
			default:
				_ = d.Skip()
			}

		case xml.EndElement:
			if el.Name.Local == "postmeta" {
				inPostMeta = false
				if currentMetaKey != "" {
					item.PostMeta[currentMetaKey] = currentMetaVal
				}
			}
			if el.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

// CleanContent sanitizes raw WordPress content and cleans up shortcodes and Gutenberg blocks.
func CleanContent(raw string, cleanShortcodes bool, sanitize bool, sanitizer *bluemonday.Policy) string {
	content := strings.TrimSpace(raw)
	if content == "" {
		return ""
	}

	if cleanShortcodes {
		// 1. Remove Gutenberg block comments
		content = reGutenbergComments.ReplaceAllLiteralString(content, "")

		// 2. Transform [caption] shortcode to semantic HTML5 <figure>
		content = reCaption.ReplaceAllStringFunc(content, func(match string) string {
			sub := reCaption.FindStringSubmatch(match)
			if len(sub) < 2 {
				return ""
			}
			inner := strings.TrimSpace(sub[1])
			img := reImgTag.FindString(inner)
			captionText := strings.TrimSpace(reImgTag.ReplaceAllString(inner, ""))
			if img != "" {
				if captionText != "" {
					return fmt.Sprintf(`<figure class="wp-caption my-4">%s<figcaption class="text-xs text-muted-foreground mt-1 text-center">%s</figcaption></figure>`, img, html.EscapeString(captionText))
				}
				return img
			}
			return inner
		})

		// 3. Transform [embed] to standard links
		content = reEmbed.ReplaceAllStringFunc(content, func(match string) string {
			sub := reEmbed.FindStringSubmatch(match)
			if len(sub) < 2 {
				return ""
			}
			rawURL := strings.TrimSpace(sub[1])
			return fmt.Sprintf(`<p><a href="%s" target="_blank" rel="noopener noreferrer">%s</a></p>`, html.EscapeString(rawURL), html.EscapeString(rawURL))
		})

		// 4. Transform audio & video shortcodes
		content = reAudioVideo.ReplaceAllStringFunc(content, func(match string) string {
			sub := reAudioVideo.FindStringSubmatch(match)
			if len(sub) >= 3 {
				tag := sub[1]
				src := sub[2]
				return fmt.Sprintf(`<%s controls src="%s" class="w-full my-4"></%s>`, tag, html.EscapeString(src), tag)
			}
			return ""
		})

		// 5. Remove [gallery] and any other leftover shortcodes
		content = reGallery.ReplaceAllLiteralString(content, "")
		content = reRemainingTags.ReplaceAllLiteralString(content, "")
	}

	// 6. Ensure paragraphs exist (classic WP autop fallback)
	if !strings.Contains(content, "<p>") && !strings.Contains(content, "<div") {
		paras := strings.Split(content, "\n\n")
		var sb strings.Builder
		for _, para := range paras {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}
			para = strings.ReplaceAll(para, "\n", "<br/>")
			sb.WriteString(fmt.Sprintf("<p class=\"mb-4 leading-relaxed\">%s</p>\n", para))
		}
		content = sb.String()
	}

	// 7. Sanitize via bluemonday policy
	if sanitize {
		if sanitizer == nil {
			sanitizer = createSanitizer()
		}
		content = sanitizer.Sanitize(content)
	}

	content = reExcessiveSpaces.ReplaceAllString(content, "\n\n")
	return strings.TrimSpace(content)
}

// Slugify generates a clean SEO-friendly slug with Cyrillic transliteration.
func Slugify(s string) string {
	var ruTranslit = map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}

	s = strings.ToLower(strings.TrimSpace(s))
	var sb strings.Builder
	for _, r := range s {
		if val, ok := ruTranslit[r]; ok {
			sb.WriteString(val)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			sb.WriteRune('-')
		}
	}
	res := sb.String()
	reg := regexp.MustCompile(`-+`)
	res = reg.ReplaceAllString(res, "-")
	res = strings.Trim(res, "-")
	if res == "" {
		res = "item-" + uuid.New().String()[:8]
	}
	return res
}

// ParseAndImport reads standard WordPress WXR XML stream and imports articles, pages, and redirects.
func (imp *Importer) ParseAndImport(ctx context.Context, r io.Reader, opts Options, onProgress ProgressCallback) (*Progress, error) {
	prog := Progress{}
	if onProgress != nil {
		onProgress(prog)
	}

	attachments := make(map[string]string) // post_id -> attachment_url
	type delayedArticle struct {
		article models.ArticleItem
		thumbID string
		origURL string
	}
	var delayedArticles []delayedArticle

	decoder := xml.NewDecoder(r)

	report := func(current string) {
		prog.CurrentItem = current
		if prog.TotalItems > 0 {
			prog.Percentage = (prog.ProcessedItems * 100) / prog.TotalItems
			if prog.Percentage > 100 {
				prog.Percentage = 100
			}
		}
		if onProgress != nil {
			onProgress(prog)
		}
	}

	for {
		if ctx.Err() != nil {
			return &prog, ctx.Err()
		}

		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Warn("WXR XML stream token error", "error", err)
			break
		}

		se, ok := token.(xml.StartElement)
		if !ok || se.Name.Local != "item" {
			continue
		}

		prog.TotalItems++
		var item WXRItem
		if err := decoder.DecodeElement(&item, &se); err != nil {
			prog.Errors = append(prog.Errors, fmt.Sprintf("Error decoding item: %v", err))
			continue
		}

		prog.ProcessedItems++

		// 1. Process Attachments
		if item.PostType == "attachment" {
			if item.PostID != "" && item.AttachmentURL != "" {
				attachments[item.PostID] = item.AttachmentURL
			}
			report("Медиафайл: " + item.Title)
			continue
		}

		// Check status filter
		isPublished := (item.Status == "publish")
		if opts.PublishStatus == "published_only" && !isPublished {
			continue
		}

		// 2. Process Blog Posts
		if item.PostType == "post" && opts.ImportPosts {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				title = "Без названия (" + item.PostID + ")"
			}

			slug := strings.TrimSpace(item.PostName)
			if slug == "" {
				slug = Slugify(title)
			}

			content := CleanContent(item.Content, opts.CleanShortcodes, opts.SanitizeHTML, imp.sanitizer)
			excerpt := strings.TrimSpace(item.Excerpt)
			if excerpt == "" {
				// Derive 180 chars excerpt from content
				plain := bluemonday.StripTagsPolicy().Sanitize(content)
				if len([]rune(plain)) > 180 {
					excerpt = string([]rune(plain)[:180]) + "..."
				} else {
					excerpt = plain
				}
			}

			author := strings.TrimSpace(item.Creator)
			if author == "" {
				author = opts.DefaultAuthor
			}

			pubDate := strings.TrimSpace(item.PostDate)
			if pubDate == "" {
				pubDate = time.Now().Format("2006-01-02 15:04:05")
			}

			var tagsList []string
			for _, cat := range item.Categories {
				catName := strings.TrimSpace(cat.Name)
				if catName != "" {
					tagsList = append(tagsList, catName)
				}
			}

			coverImage := ""
			thumbID := item.PostMeta["_thumbnail_id"]
			if thumbID != "" && attachments[thumbID] != "" {
				coverImage = attachments[thumbID]
			} else {
				// Extract first <img src="..."> as fallback cover image
				if imgMatch := reImgTag.FindString(item.Content); imgMatch != "" {
					if srcMatch := regexp.MustCompile(`src=["']([^"']+)["']`).FindStringSubmatch(imgMatch); len(srcMatch) > 1 {
						coverImage = srcMatch[1]
					}
				}
			}

			art := models.ArticleItem{
				ID:             uuid.New().String(),
				Title:          title,
				Slug:           slug,
				Excerpt:        excerpt,
				Content:        content,
				CoverImage:     coverImage,
				AuthorName:     author,
				PublishedAt:    pubDate,
				IsPublished:    isPublished,
				SEOTitle:       title,
				SEODescription: excerpt,
				Tags:           tagsList,
			}

			// If thumbID exists but attachment hasn't appeared yet, buffer for post-processing
			if thumbID != "" && coverImage == "" {
				delayedArticles = append(delayedArticles, delayedArticle{
					article: art,
					thumbID: thumbID,
					origURL: item.Link,
				})
			} else {
				if err := imp.saveArticleAndRedirect(ctx, art, item.Link, opts.CreateRedirects, &prog); err != nil {
					prog.Errors = append(prog.Errors, fmt.Sprintf("Article save '%s': %v", title, err))
				}
			}
			report("Статья: " + title)
			continue
		}

		// 3. Process Pages
		if item.PostType == "page" && opts.ImportPages {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				title = "Страница (" + item.PostID + ")"
			}
			slug := strings.TrimSpace(item.PostName)
			if slug == "" {
				slug = Slugify(title)
			}

			content := CleanContent(item.Content, opts.CleanShortcodes, opts.SanitizeHTML, imp.sanitizer)

			if err := imp.savePageAndRedirect(ctx, title, slug, content, item.Link, isPublished, opts.CreateRedirects, &prog); err != nil {
				prog.Errors = append(prog.Errors, fmt.Sprintf("Page save '%s': %v", title, err))
			}
			report("Страница: " + title)
			continue
		}
	}

	// Post-process delayed articles whose attachment metadata was parsed later in stream
	for _, da := range delayedArticles {
		if da.article.CoverImage == "" && da.thumbID != "" && attachments[da.thumbID] != "" {
			da.article.CoverImage = attachments[da.thumbID]
		}
		if err := imp.saveArticleAndRedirect(ctx, da.article, da.origURL, opts.CreateRedirects, &prog); err != nil {
			prog.Errors = append(prog.Errors, fmt.Sprintf("Delayed article save '%s': %v", da.article.Title, err))
		}
	}

	// Invalidate caches once import has finalized
	_, _ = imp.repo.GetRedirectsMap(ctx)

	prog.IsComplete = true
	prog.Percentage = 100
	prog.CurrentItem = fmt.Sprintf("Завершено! Статей: %d, Страниц: %d, 301-редиректов: %d", prog.PostsImported, prog.PagesImported, prog.RedirectsCreated)
	if onProgress != nil {
		onProgress(prog)
	}

	return &prog, nil
}

func (imp *Importer) saveArticleAndRedirect(ctx context.Context, art models.ArticleItem, origURL string, createRedirect bool, prog *Progress) error {
	// Persist article
	if err := imp.repo.SaveArticle(ctx, art); err != nil {
		return err
	}
	prog.PostsImported++

	if !createRedirect {
		return nil
	}

	// 301 Redirect preservation
	targetPath := "/blog/" + art.Slug
	imp.registerRedirects(ctx, origURL, targetPath, prog)
	return nil
}

func (imp *Importer) savePageAndRedirect(ctx context.Context, title, slug, content, origURL string, isPublished bool, createRedirect bool, prog *Progress) error {
	pubInt := int64(1)
	if !isPublished {
		pubInt = 0
	}

	// Check if page exists or create
	_, err := imp.repo.CreatePage(ctx, slug, title, title, "", pubInt)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "unique") {
		slog.Debug("page creation skipped or failed", "slug", slug, "error", err)
	}

	// Also create a PageBlock for the rich text content
	if content != "" {
		_, _ = imp.repo.CreatePageBlock(ctx, models.PageBlock{
			PageID:  slug,
			Type:    "rich_text",
			Title:   title,
			Content: content,
			Order:   10,
			Active:  true,
		})
	}
	prog.PagesImported++

	if !createRedirect {
		return nil
	}

	targetPath := "/" + slug
	if slug == "home" || slug == "glavnaya" {
		targetPath = "/"
	}
	imp.registerRedirects(ctx, origURL, targetPath, prog)
	return nil
}

func (imp *Importer) registerRedirects(ctx context.Context, origURL, targetPath string, prog *Progress) {
	if origURL == "" {
		return
	}

	// Extract path from origURL
	sourcePath := origURL
	if u, err := url.Parse(origURL); err == nil && u.Path != "" {
		sourcePath = u.RequestURI()
	}

	sourcePath = data.NormalizeRedirectPath(sourcePath)
	targetPath = data.NormalizeRedirectPath(targetPath)

	if sourcePath == "" || targetPath == "" || sourcePath == targetPath {
		return
	}

	// Upsert redirect in DB
	_, err := imp.repo.UpsertRedirect(ctx, models.RedirectRule{
		SourcePath: sourcePath,
		TargetPath: targetPath,
		StatusCode: 301,
	})
	if err == nil {
		prog.RedirectsCreated++
	}

	// Also register without trailing slash if different and has no query
	if !strings.Contains(sourcePath, "?") {
		trimmed := strings.TrimSuffix(sourcePath, "/")
		if trimmed != "" && trimmed != sourcePath && trimmed != targetPath {
			_, _ = imp.repo.UpsertRedirect(ctx, models.RedirectRule{
				SourcePath: trimmed,
				TargetPath: targetPath,
				StatusCode: 301,
			})
		}
	}
}

// JobManager tracks asynchronous import tasks in memory for SSE streaming.
type JobManager struct {
	mu   sync.RWMutex
	jobs map[string]*ImportJob
}

// ImportJob represents a background import operation.
type ImportJob struct {
	ID        string
	Progress  Progress
	Updates   chan Progress
	CreatedAt time.Time
}

var DefaultJobManager = &JobManager{
	jobs: make(map[string]*ImportJob),
}

// CreateJob creates a new tracked import job.
func (jm *JobManager) CreateJob(id string) *ImportJob {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	job := &ImportJob{
		ID:        id,
		Updates:   make(chan Progress, 100),
		CreatedAt: time.Now(),
	}
	jm.jobs[id] = job
	return job
}

// GetJob retrieves a job by ID.
func (jm *JobManager) GetJob(id string) (*ImportJob, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	job, ok := jm.jobs[id]
	return job, ok
}

// CleanOldJobs prunes finished jobs older than 1 hour.
func (jm *JobManager) CleanOldJobs() {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	now := time.Now()
	for id, job := range jm.jobs {
		if now.Sub(job.CreatedAt) > time.Hour {
			delete(jm.jobs, id)
		}
	}
}
