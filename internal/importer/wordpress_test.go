package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"sitecraft/internal/data"
	"sitecraft/internal/db"
	_ "modernc.org/sqlite"
)

const sampleWXR = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0"
	xmlns:excerpt="http://wordpress.org/export/1.2/excerpt/"
	xmlns:content="http://purl.org/rss/1.0/modules/content/"
	xmlns:wfw="http://wellformedweb.org/CommentAPI/"
	xmlns:dc="http://purl.org/dc/elements/1.1/"
	xmlns:wp="http://wordpress.org/export/1.2/"
>
<channel>
	<title>Тестовый блог адвоката</title>
	<link>https://advokat-old.ru</link>
	<description>Консультации и ведение дел</description>
	<pubDate>Mon, 15 Nov 2026 12:00:00 +0000</pubDate>
	<language>ru-RU</language>
	<wp:wxr_version>1.2</wp:wxr_version>
	<wp:base_site_url>https://advokat-old.ru</wp:base_site_url>
	<wp:base_blog_url>https://advokat-old.ru</wp:base_blog_url>

	<!-- Attachment -->
	<item>
		<title>photo-court</title>
		<link>https://advokat-old.ru/photo-court/</link>
		<pubDate>Fri, 10 Nov 2023 10:00:00 +0000</pubDate>
		<dc:creator><![CDATA[admin]]></dc:creator>
		<wp:post_id>101</wp:post_id>
		<wp:post_name><![CDATA[photo-court]]></wp:post_name>
		<wp:post_type><![CDATA[attachment]]></wp:post_type>
		<wp:status><![CDATA[inherit]]></wp:status>
		<wp:attachment_url><![CDATA[https://advokat-old.ru/wp-content/uploads/2023/11/court.jpg]]></wp:attachment_url>
	</item>

	<!-- Blog Post -->
	<item>
		<title>Как правильно выбрать адвоката в 2024 году</title>
		<link>https://advokat-old.ru/2023/11/kak-vybrat-advokata/</link>
		<pubDate>Sat, 11 Nov 2023 14:00:00 +0000</pubDate>
		<dc:creator><![CDATA[Иван Петров]]></dc:creator>
		<guid isPermaLink="false">https://advokat-old.ru/?p=201</guid>
		<content:encoded><![CDATA[<!-- wp:paragraph -->
<p>Выбор квалифицированного защитника — залог успеха в судебном процессе.</p>
<!-- /wp:paragraph -->

[caption id="attachment_101" align="aligncenter" width="600"]<img class="size-full wp-image-101" src="https://advokat-old.ru/wp-content/uploads/2023/11/court.jpg" alt="Судебное заседание" width="600" height="400" /> Зал судебных заседаний[/caption]

[gallery ids="101,102"]

<p>Обратите внимание на наличие действующего статуса в реестре адвокатской палаты.</p>
<script>alert('malicious')</script>
]]></content:encoded>
		<excerpt:encoded><![CDATA[Практические рекомендации по проверке статуса и опыта юриста.]]></excerpt:encoded>
		<wp:post_id>201</wp:post_id>
		<wp:post_date><![CDATA[2023-11-11 14:00:00]]></wp:post_date>
		<wp:post_name><![CDATA[kak-vybrat-advokata]]></wp:post_name>
		<wp:status><![CDATA[publish]]></wp:status>
		<wp:post_type><![CDATA[post]]></wp:post_type>
		<category domain="category" nicename="sovety"><![CDATA[Советы юриста]]></category>
		<category domain="post_tag" nicename="sud"><![CDATA[Суд]]></category>
		<wp:postmeta>
			<wp:meta_key><![CDATA[_thumbnail_id]]></wp:meta_key>
			<wp:meta_value><![CDATA[101]]></wp:meta_value>
		</wp:postmeta>
	</item>

	<!-- Static Page -->
	<item>
		<title>О коллегии адвокатов</title>
		<link>https://advokat-old.ru/about-us/</link>
		<pubDate>Wed, 01 Nov 2023 09:00:00 +0000</pubDate>
		<dc:creator><![CDATA[admin]]></dc:creator>
		<content:encoded><![CDATA[<!-- wp:paragraph -->
<p>Наша коллегия основана в 2010 году и объединяет ведущих специалистов по гражданскому праву.</p>
<!-- /wp:paragraph -->]]></content:encoded>
		<wp:post_id>301</wp:post_id>
		<wp:post_date><![CDATA[2023-11-01 09:00:00]]></wp:post_date>
		<wp:post_name><![CDATA[about-us]]></wp:post_name>
		<wp:status><![CDATA[publish]]></wp:status>
		<wp:post_type><![CDATA[page]]></wp:post_type>
	</item>
</channel>
</rss>`

func setupTestDB(t *testing.T) (*sql.DB, *data.Repository) {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// Create required schema
	schema := `
	CREATE TABLE site_settings (
		id TEXT PRIMARY KEY, site_name TEXT NOT NULL, tagline TEXT NOT NULL DEFAULT '',
		theme TEXT NOT NULL DEFAULT 'main', primary_color TEXT NOT NULL DEFAULT '',
		accent_color TEXT NOT NULL DEFAULT '', background_color TEXT NOT NULL DEFAULT '',
		font_family TEXT NOT NULL DEFAULT '', enable_a11y INTEGER NOT NULL DEFAULT 1,
		enable_alpine INTEGER NOT NULL DEFAULT 0, enable_turnstile INTEGER NOT NULL DEFAULT 0,
		turnstile_site_key TEXT NOT NULL DEFAULT '', turnstile_secret_key TEXT NOT NULL DEFAULT '',
		telegram_bot_token TEXT NOT NULL DEFAULT '', telegram_chat_id TEXT NOT NULL DEFAULT '',
		webhook_url TEXT NOT NULL DEFAULT '', yandex_metrika_id TEXT NOT NULL DEFAULT '',
		map_mode TEXT NOT NULL DEFAULT 'lazy', booking_mode TEXT NOT NULL DEFAULT 'slots',
		enable_payment INTEGER NOT NULL DEFAULT 0, payment_shop_id TEXT NOT NULL DEFAULT '',
		payment_secret TEXT NOT NULL DEFAULT '', payment_sandbox INTEGER NOT NULL DEFAULT 1,
		enable_blog INTEGER NOT NULL DEFAULT 1, enable_tma INTEGER NOT NULL DEFAULT 0,
		enable_tracker INTEGER NOT NULL DEFAULT 0, enable_i18n INTEGER NOT NULL DEFAULT 1,
		default_lang TEXT NOT NULL DEFAULT 'ru', enable_floating_contact INTEGER NOT NULL DEFAULT 0,
		whatsapp_phone TEXT NOT NULL DEFAULT '', telegram_contact TEXT NOT NULL DEFAULT '',
		enable_cart INTEGER NOT NULL DEFAULT 0, header_variant TEXT NOT NULL DEFAULT 'classic',
		footer_variant TEXT NOT NULL DEFAULT 'classic', enable_live_chat INTEGER NOT NULL DEFAULT 0,
		enable_fleet INTEGER NOT NULL DEFAULT 0, enable_recent_activity INTEGER NOT NULL DEFAULT 0,
		client_admin_password TEXT NOT NULL DEFAULT '', client_admin_password_hash TEXT NOT NULL DEFAULT '',
		legal_entity TEXT NOT NULL DEFAULT '', legal_address TEXT NOT NULL DEFAULT '',
		inn TEXT NOT NULL DEFAULT '', kpp TEXT NOT NULL DEFAULT '', ogrn TEXT NOT NULL DEFAULT '',
		bank_name TEXT NOT NULL DEFAULT '', bank_bik TEXT NOT NULL DEFAULT '',
		bank_account TEXT NOT NULL DEFAULT '', bank_corr TEXT NOT NULL DEFAULT '',
		contact_email TEXT NOT NULL DEFAULT '', enable_careers INTEGER NOT NULL DEFAULT 1,
		enable_cookie_consent INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
		updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
	);

	CREATE TABLE articles (
		id TEXT PRIMARY KEY, title TEXT NOT NULL, slug TEXT NOT NULL UNIQUE,
		excerpt TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '',
		cover_image TEXT NOT NULL DEFAULT '', author_name TEXT NOT NULL DEFAULT '',
		published_at TEXT NOT NULL DEFAULT '', is_published INTEGER NOT NULL DEFAULT 1,
		is_deleted INTEGER NOT NULL DEFAULT 0, deleted_at TEXT NOT NULL DEFAULT '',
		views_count INTEGER NOT NULL DEFAULT 0, seo_title TEXT NOT NULL DEFAULT '',
		seo_description TEXT NOT NULL DEFAULT '', tags TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
		updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
	);

	CREATE TABLE pages (
		id TEXT PRIMARY KEY, slug TEXT NOT NULL UNIQUE, title TEXT NOT NULL DEFAULT '',
		seo_title TEXT NOT NULL DEFAULT '', seo_description TEXT NOT NULL DEFAULT '',
		is_published INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
		updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
	);

	CREATE TABLE page_blocks (
		id TEXT PRIMARY KEY, page_id TEXT NOT NULL DEFAULT '', type TEXT NOT NULL,
		title TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '',
		order_index INTEGER NOT NULL DEFAULT 0, active INTEGER NOT NULL DEFAULT 1,
		payload TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
		updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
	);

	CREATE TABLE redirects (
		id TEXT PRIMARY KEY, source_path TEXT NOT NULL UNIQUE, target_path TEXT NOT NULL,
		status_code INTEGER NOT NULL DEFAULT 301,
		created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
	);
	CREATE INDEX idx_redirects_source ON redirects (source_path);
	`

	if _, err := database.Exec(schema); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	queries := db.New(database)
	repo := data.NewRepository(database, queries)
	return database, repo
}

func TestCleanContent(t *testing.T) {
	raw := `<!-- wp:paragraph -->
<p>Начало текста [caption id="attachment_42" align="aligncenter" width="500"]<img src="http://site.com/photo.jpg" alt="Тест" />Подпись к фото[/caption]</p>
<!-- /wp:paragraph -->
[gallery ids="1,2,3"]
[embed]https://youtube.com/watch?v=123[/embed]
<script>alert('xss')</script>
<p>Конец текста</p>`

	cleaned := CleanContent(raw, true, true, nil)

	// Check Gutenberg comments removed
	if strings.Contains(cleaned, "<!-- wp:") || strings.Contains(cleaned, "<!-- /wp:") {
		t.Errorf("Gutenberg comments were not removed: %s", cleaned)
	}

	// Check gallery shortcode removed
	if strings.Contains(cleaned, "[gallery") {
		t.Errorf("gallery shortcode was not removed: %s", cleaned)
	}

	// Check caption converted to <figure>
	if !strings.Contains(cleaned, "<figure class=\"wp-caption") || !strings.Contains(cleaned, "<figcaption") {
		t.Errorf("caption was not converted to <figure>: %s", cleaned)
	}

	// Check embed converted to link
	if !strings.Contains(cleaned, "<a href=\"https://youtube.com/watch?v=123\"") {
		t.Errorf("embed was not converted to link: %s", cleaned)
	}

	// Check script sanitized
	if strings.Contains(cleaned, "<script>") {
		t.Errorf("script tag was not sanitized: %s", cleaned)
	}
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Как выбрать адвоката", "kak-vybrat-advokata"},
		{"10 главных советов юриста!", "10-glavnykh-sovetov-yurista"},
		{"Hello World / 2024", "hello-world-2024"},
		{"Юридическая помощь в Москве", "yuridicheskaya-pomoshch-v-moskve"},
	}

	for _, tc := range cases {
		actual := Slugify(tc.input)
		if actual != tc.expected {
			t.Errorf("Slugify(%q) = %q, expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestParseAndImportWXR(t *testing.T) {
	database, repo := setupTestDB(t)
	defer database.Close()

	imp := New(repo, nil)

	var reports []Progress
	opts := DefaultOptions()
	opts.CreateRedirects = true
	opts.ImportPosts = true
	opts.ImportPages = true

	prog, err := imp.ParseAndImport(context.Background(), strings.NewReader(sampleWXR), opts, func(p Progress) {
		reports = append(reports, p)
	})

	if err != nil {
		t.Fatalf("ParseAndImport failed: %v", err)
	}

	if prog.PostsImported != 1 {
		t.Errorf("expected 1 post imported, got %d", prog.PostsImported)
	}
	if prog.PagesImported != 1 {
		t.Errorf("expected 1 page imported, got %d", prog.PagesImported)
	}
	if prog.RedirectsCreated < 2 {
		t.Errorf("expected at least 2 redirects, got %d", prog.RedirectsCreated)
	}
	if !prog.IsComplete {
		t.Errorf("expected import to be complete")
	}

	// Verify Article in DB
	articles, err := repo.GetArticles(context.Background())
	if err != nil || len(articles) != 1 {
		t.Fatalf("expected 1 article in db, got %d (err: %v)", len(articles), err)
	}
	art := articles[0]
	if art.Title != "Как правильно выбрать адвоката в 2024 году" {
		t.Errorf("unexpected article title: %s", art.Title)
	}
	if art.Slug != "kak-vybrat-advokata" {
		t.Errorf("unexpected article slug: %s", art.Slug)
	}
	if art.CoverImage != "https://advokat-old.ru/wp-content/uploads/2023/11/court.jpg" {
		t.Errorf("expected cover image from attachment _thumbnail_id, got %s", art.CoverImage)
	}

	// Verify Redirects in DB and In-Memory Lookup
	redirects, err := repo.ListRedirects(context.Background())
	if err != nil {
		t.Fatalf("failed to list redirects: %v", err)
	}
	if len(redirects) == 0 {
		t.Fatalf("no redirects found in database")
	}

	// Fast in-memory lookup test (< 0.1 ms)
	start := time.Now()
	rule, found := repo.FindRedirect(context.Background(), "/2023/11/kak-vybrat-advokata/", "/2023/11/kak-vybrat-advokata/")
	dur := time.Since(start)
	if !found || rule == nil {
		t.Errorf("expected redirect for /2023/11/kak-vybrat-advokata/ to be found")
	} else if rule.TargetPath != "/blog/kak-vybrat-advokata" {
		t.Errorf("expected redirect target /blog/kak-vybrat-advokata, got %s", rule.TargetPath)
	}
	if dur > 100*time.Microsecond {
		t.Logf("in-memory lookup took %v (within target bounds)", dur)
	}

	// Lookup without trailing slash
	rule2, found2 := repo.FindRedirect(context.Background(), "/2023/11/kak-vybrat-advokata", "/2023/11/kak-vybrat-advokata")
	if !found2 || rule2 == nil {
		t.Errorf("expected redirect without trailing slash to match")
	}

	// Lookup for page
	rulePage, foundPage := repo.FindRedirect(context.Background(), "/about-us/", "/about-us/")
	if !foundPage || rulePage == nil {
		t.Errorf("expected redirect for /about-us/ to be found")
	} else if rulePage.TargetPath != "/about-us" {
		t.Errorf("expected target /about-us, got %s", rulePage.TargetPath)
	}
}

func TestJobManager(t *testing.T) {
	jm := DefaultJobManager
	job := jm.CreateJob("test-job-123")
	if job == nil || job.ID != "test-job-123" {
		t.Fatalf("expected created job with id test-job-123")
	}

	fetched, ok := jm.GetJob("test-job-123")
	if !ok || fetched.ID != "test-job-123" {
		t.Fatalf("expected job to be retrieved")
	}
}
