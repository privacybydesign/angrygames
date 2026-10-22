package main

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testAuthToken   = "secret-token"
	testIrmaSession = `{"sessionPtr":{"u":"http://irma.example/irma/abc","irmaqr":"disclosing"},"token":"tok"}`
	testIrmaURL     = "http://irma.invalid"
	testStart18Plus = "/start-session?type=18plus"

	testInvalidRequest = "Invalid request"
	testServerError    = "Internal server error"

	testHTMLType = "text/html; charset=utf-8"
	testJSType   = "text/javascript; charset=utf-8"
)

// Wire values are spelled out here instead of reusing the server's constants,
// so a wrong constant there fails a test.
const (
	wireAllow         = "Allow"
	wireAuthorization = "Authorization"
	wireCacheControl  = "Cache-Control"
	wireContentType   = "Content-Type"
	wireJSON          = "application/json"
	wireNoStore       = "no-store"
)

// fakeIrma records the last request it received and answers with status/body.
type fakeIrma struct {
	*httptest.Server
	path          string
	authorization string
	contentType   string
	body          string
}

func newFakeIrma(t *testing.T, status int, body string) *fakeIrma {
	t.Helper()

	f := &fakeIrma{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ := io.ReadAll(r.Body)
		f.path = r.Method + " " + r.URL.Path
		f.authorization = r.Header.Get(wireAuthorization)
		f.contentType = r.Header.Get(wireContentType)
		f.body = string(received)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(f.Close)
	return f
}

func get(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestStartSessionAgeTypes(t *testing.T) {
	for _, age := range []string{"12", "16", "18"} {
		t.Run(age+"plus", func(t *testing.T) {
			irma := newFakeIrma(t, http.StatusOK, testIrmaSession)
			h := newHandler(newIrmaClient(irma.URL, testAuthToken))

			rec := get(t, h, "/start-session?type="+age+"plus")

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if rec.Body.String() != testIrmaSession {
				t.Errorf("body = %q, want the IRMA server response unchanged", rec.Body.String())
			}
			if got := rec.Header().Get(wireContentType); got != wireJSON {
				t.Errorf("Content-Type = %q, want %q", got, wireJSON)
			}
			if got := rec.Header().Get(wireCacheControl); got != wireNoStore {
				t.Errorf("Cache-Control = %q, want %q", got, wireNoStore)
			}
			if irma.path != "POST /session" {
				t.Errorf("IRMA server got %q, want POST /session", irma.path)
			}
			if irma.authorization != testAuthToken {
				t.Errorf("Authorization = %q, want %q", irma.authorization, testAuthToken)
			}
			if irma.contentType != wireJSON {
				t.Errorf("Content-Type to IRMA = %q, want %q", irma.contentType, wireJSON)
			}

			want := `{"@context":"https://irma.app/ld/request/disclosure/v2","disclose":[[` +
				`["pbdf.pbdf.ageLimits.over` + age + `"],` +
				`["pbdf.nijmegen.ageLimits.over` + age + `"],` +
				`["pbdf.gemeente.personalData.over` + age + `"],` +
				`["pbdf.pilot-amsterdam.idcard.over` + age + `"],` +
				`["pbdf.pilot-amsterdam.passport.over` + age + `"]]]}`
			if irma.body != want {
				t.Errorf("session request =\n%s\nwant\n%s", irma.body, want)
			}
		})
	}
}

func TestStartSessionWithoutToken(t *testing.T) {
	irma := newFakeIrma(t, http.StatusOK, testIrmaSession)
	h := newHandler(newIrmaClient(irma.URL, ""))

	rec := get(t, h, testStart18Plus)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if irma.authorization != "" {
		t.Errorf("Authorization = %q, want none", irma.authorization)
	}
}

func TestStartSessionInvalidType(t *testing.T) {
	for _, target := range []string{
		"/start-session",
		"/start-session?type=",
		"/start-session?type=21plus",
		"/start-session?type=18",
		"/start-session?type=18PLUS",
	} {
		t.Run(target, func(t *testing.T) {
			irma := newFakeIrma(t, http.StatusOK, testIrmaSession)
			h := newHandler(newIrmaClient(irma.URL, testAuthToken))

			rec := get(t, h, target)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != testInvalidRequest {
				t.Errorf("body = %q, want %q", got, testInvalidRequest)
			}
			if irma.path != "" {
				t.Errorf("IRMA server was called with %q, want no call", irma.path)
			}
		})
	}
}

func TestStartSessionIrmaFailure(t *testing.T) {
	unreachable := newFakeIrma(t, http.StatusOK, "")
	unreachable.Close()

	tests := map[string]string{
		"server error": newFakeIrma(t, http.StatusInternalServerError, "boom").URL,
		"unauthorized": newFakeIrma(t, http.StatusUnauthorized, "no").URL,
		"empty body":   newFakeIrma(t, http.StatusOK, "").URL,
		"unreachable":  unreachable.URL,
		"invalid url":  "://bad",
		"empty url":    "",
	}

	for name, url := range tests {
		t.Run(name, func(t *testing.T) {
			h := newHandler(newIrmaClient(url, testAuthToken))

			rec := get(t, h, testStart18Plus)

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want 500", rec.Code)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != testServerError {
				t.Errorf("body = %q, want %q", got, testServerError)
			}
		})
	}
}

func TestStartSessionRejectsPost(t *testing.T) {
	irma := newFakeIrma(t, http.StatusOK, testIrmaSession)
	h := newHandler(newIrmaClient(irma.URL, testAuthToken))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, testStart18Plus, nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
	if irma.path != "" {
		t.Errorf("IRMA server was called with %q, want no call", irma.path)
	}
}

func TestStartSessionRejectsHead(t *testing.T) {
	irma := newFakeIrma(t, http.StatusOK, testIrmaSession)
	h := newHandler(newIrmaClient(irma.URL, testAuthToken))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, testStart18Plus, nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
	if got := rec.Header().Get(wireAllow); got != http.MethodGet {
		t.Errorf("Allow = %q, want GET", got)
	}
	if irma.path != "" {
		t.Errorf("IRMA server was called with %q, want no call", irma.path)
	}
}

func TestSessionRequestsMarshal(t *testing.T) {
	for name, request := range sessionRequests {
		if _, err := json.Marshal(request); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestPagesAreServed(t *testing.T) {
	h := newHandler(newIrmaClient(testIrmaURL, ""))

	for target, wantType := range map[string]string{
		"/":                  testHTMLType,
		"/cart.html":         testHTMLType,
		"/css/style.css":     "text/css; charset=utf-8",
		"/js/main.js":        testJSType,
		"/js/yivi.js":        testJSType,
		"/img/yivi-logo.svg": "image/svg+xml",
		"/img/product01.png": "image/png",
	} {
		t.Run(target, func(t *testing.T) {
			rec := get(t, h, target)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get(wireContentType); got != wantType {
				t.Errorf("Content-Type = %q, want %q", got, wantType)
			}
			if rec.Body.Len() == 0 {
				t.Error("empty body")
			}
		})
	}
}

func TestFontMimeTypes(t *testing.T) {
	h := newHandler(newIrmaClient(testIrmaURL, ""))

	for target, wantType := range map[string]string{
		"/fonts/fontawesome-webfont.eot":   "application/vnd.ms-fontobject",
		"/fonts/FontAwesome.otf":           "font/otf",
		"/fonts/fontawesome-webfont.ttf":   "font/ttf",
		"/fonts/fontawesome-webfont.woff":  "font/woff",
		"/fonts/fontawesome-webfont.woff2": "font/woff2",
		"/fonts/fontawesome-webfont.svg":   "image/svg+xml",
	} {
		t.Run(target, func(t *testing.T) {
			rec := get(t, h, target)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get(wireContentType); got != wantType {
				t.Errorf("Content-Type = %q, want %q", got, wantType)
			}
		})
	}
}

func TestNonPublicFilesNotServed(t *testing.T) {
	h := newHandler(newIrmaClient(testIrmaURL, ""))

	for _, target := range []string{
		"/config.php",
		"/start_session.php",
		"/Dockerfile",
		"/.dockerignore",
		"/.git/config",
		"/.github/workflows/delivery.yml",
		"/go.mod",
		"/main.go",
		"/server.go",
		"/server_test.go",
		"/README.md",
		"/CLAUDE.md",
		"/Credits.txt",
		"/../go.mod",
		"/css/../go.mod",
		"/%2e%2e/go.mod",
	} {
		t.Run(target, func(t *testing.T) {
			rec := get(t, h, target)

			if rec.Code == http.StatusOK {
				t.Errorf("status = 200, want the file to be unreachable")
			}
		})
	}
}

func TestEmbeddedFilesPublicOnly(t *testing.T) {
	allowedTopLevel := map[string]bool{
		"index.html": true,
		"cart.html":  true,
		"css":        true,
		"fonts":      true,
		"img":        true,
		"js":         true,
	}

	entries, err := fs.ReadDir(staticFiles, ".")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if !allowedTopLevel[entry.Name()] {
			t.Errorf("%q is embedded but is not a public asset", entry.Name())
		}
	}

	if len(entries) != len(allowedTopLevel) {
		t.Errorf("embedded %d top-level entries, want %d", len(entries), len(allowedTopLevel))
	}
}

func TestFrontendSessionPath(t *testing.T) {
	main, err := fs.ReadFile(staticFiles, "js/main.js")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(main), "`start-session?type=`") {
		t.Error("js/main.js does not start sessions at start-session")
	}
}
