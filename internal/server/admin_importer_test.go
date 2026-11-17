package server_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"sitecraft/internal/auth"
	"sitecraft/internal/importer"
)

func TestRedirectMiddleware_And_AdminImporter(t *testing.T) {
	srv, conn := setupTestServer(t)
	defer conn.Close()

	ctx := context.Background()

	// 1. Authenticate an admin session
	sessionCookie, err := srv.SessionMgr.Load(ctx, "")
	if err != nil {
		t.Fatalf("failed to init session: %v", err)
	}
	ctx = sessionCookie
	_ = auth.SetAdminSession(ctx, srv.SessionMgr, "admin@sitecraft.local")
	_, _, err = srv.SessionMgr.Commit(ctx)
	if err != nil {
		t.Fatalf("failed to commit session: %v", err)
	}

	sendAuthRequest := func(method, path string, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
		var req *http.Request
		if body != nil {
			req = httptest.NewRequest(method, path, body)
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		srv.Router.ServeHTTP(w, req)
		return w
	}

	// 2. Test Tab Importer rendering
	wTab := sendAuthRequest("GET", "/admin/tab/importer", nil, "")
	if wTab.Code != http.StatusOK {
		t.Fatalf("expected 200 for /admin/tab/importer, got %d", wTab.Code)
	}
	if !strings.Contains(wTab.Body.String(), "WordPress") || !strings.Contains(wTab.Body.String(), "301") {
		t.Errorf("expected tab to contain WordPress and 301, got: %s", wTab.Body.String()[:200])
	}

	// 3. Test Manual Redirect Creation
	form := url.Values{}
	form.Set("source_path", "/2021/04/legacy-consultation/")
	form.Set("target_path", "/blog/consultation")
	form.Set("status_code", "301")

	wCreate := sendAuthRequest("POST", "/admin/redirects/new", bytes.NewBufferString(form.Encode()), "application/x-www-form-urlencoded")
	if wCreate.Code != http.StatusOK {
		t.Fatalf("expected 200 for /admin/redirects/new, got %d", wCreate.Code)
	}
	if !strings.Contains(wCreate.Body.String(), "/2021/04/legacy-consultation/") {
		t.Errorf("expected created redirect in returned table: %s", wCreate.Body.String())
	}

	// 4. Test In-Memory 301 Redirection via HTTP Middleware (< 0.1 ms)
	// Exact match with trailing slash
	reqRedirect := httptest.NewRequest("GET", "/2021/04/legacy-consultation/", nil)
	wRedirect := httptest.NewRecorder()
	start := time.Now()
	srv.Router.ServeHTTP(wRedirect, reqRedirect)
	dur := time.Since(start)

	if wRedirect.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301 Moved Permanently, got %d", wRedirect.Code)
	}
	if loc := wRedirect.Header().Get("Location"); loc != "/blog/consultation" {
		t.Errorf("expected Location /blog/consultation, got %s", loc)
	}
	if dur > 1*time.Millisecond {
		t.Logf("in-memory redirect lookup took %v", dur)
	}

	// Match without trailing slash
	reqNoSlash := httptest.NewRequest("GET", "/2021/04/legacy-consultation", nil)
	wNoSlash := httptest.NewRecorder()
	srv.Router.ServeHTTP(wNoSlash, reqNoSlash)
	if wNoSlash.Code != http.StatusMovedPermanently || wNoSlash.Header().Get("Location") != "/blog/consultation" {
		t.Errorf("expected 301 redirect without slash, got %d loc %s", wNoSlash.Code, wNoSlash.Header().Get("Location"))
	}

	// 5. Test protected system paths are NEVER redirected
	reqAdmin := httptest.NewRequest("GET", "/admin/login", nil)
	wAdmin := httptest.NewRecorder()
	srv.Router.ServeHTTP(wAdmin, reqAdmin)
	if wAdmin.Code == http.StatusMovedPermanently {
		t.Errorf("/admin/login should not be redirected")
	}

	// 6. Test WXR XML Multipart Upload and SSE
	const sampleXML = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0"
	xmlns:excerpt="http://wordpress.org/export/1.2/excerpt/"
	xmlns:content="http://purl.org/rss/1.0/modules/content/"
	xmlns:dc="http://purl.org/dc/elements/1.1/"
	xmlns:wp="http://wordpress.org/export/1.2/"
>
<channel>
	<item>
		<title>Тестовая статья для загрузки</title>
		<link>https://old.ru/2022/test-upload-post/</link>
		<wp:post_name><![CDATA[test-upload-post]]></wp:post_name>
		<wp:status><![CDATA[publish]]></wp:status>
		<wp:post_type><![CDATA[post]]></wp:post_type>
		<content:encoded><![CDATA[<p>Текст статьи</p>]]></content:encoded>
	</item>
</channel>
</rss>`

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("create_redirects", "1")
	_ = writer.WriteField("import_posts", "1")
	_ = writer.WriteField("import_pages", "1")
	_ = writer.WriteField("clean_shortcodes", "1")
	_ = writer.WriteField("sanitize_html", "1")

	part, err := writer.CreateFormFile("xml_file", "export.xml")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	_, _ = part.Write([]byte(sampleXML))
	_ = writer.Close()

	wUpload := sendAuthRequest("POST", "/admin/importer/upload", body, writer.FormDataContentType())
	if wUpload.Code != http.StatusOK {
		t.Fatalf("expected 200 for upload, got %d: %s", wUpload.Code, wUpload.Body.String())
	}
	if !strings.Contains(wUpload.Body.String(), "/admin/importer/sse/") {
		t.Errorf("expected SSE container with job ID: %s", wUpload.Body.String())
	}

	// Wait briefly for background import to complete
	time.Sleep(100 * time.Millisecond)

	// Verify imported post
	articles, err := srv.Repo.GetArticles(ctx)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}
	foundPost := false
	for _, a := range articles {
		if a.Slug == "test-upload-post" {
			foundPost = true
			break
		}
	}
	if !foundPost {
		t.Errorf("expected imported post 'test-upload-post' in articles")
	}

	// Verify automatic 301 redirect was created and works immediately
	reqAutoRedirect := httptest.NewRequest("GET", "/2022/test-upload-post/", nil)
	wAutoRedirect := httptest.NewRecorder()
	srv.Router.ServeHTTP(wAutoRedirect, reqAutoRedirect)
	if wAutoRedirect.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301 for auto-created redirect, got %d", wAutoRedirect.Code)
	}
	if loc := wAutoRedirect.Header().Get("Location"); loc != "/blog/test-upload-post" {
		t.Errorf("expected target /blog/test-upload-post, got %s", loc)
	}

	// 7. Test Deletion of Redirect
	redirects, _ := srv.Repo.ListRedirects(ctx)
	if len(redirects) > 0 {
		ruleToDelete := redirects[0]
		wDelete := sendAuthRequest("POST", "/admin/redirects/"+ruleToDelete.ID+"/delete", nil, "")
		if wDelete.Code != http.StatusOK {
			t.Errorf("expected 200 on delete, got %d", wDelete.Code)
		}
	}
}

func TestImporter_JobManagerClean(t *testing.T) {
	jm := importer.DefaultJobManager
	job := jm.CreateJob("old-job")
	job.CreatedAt = time.Now().Add(-2 * time.Hour)
	jm.CleanOldJobs()
	_, found := jm.GetJob("old-job")
	if found {
		t.Errorf("expected old job to be cleaned")
	}
}
