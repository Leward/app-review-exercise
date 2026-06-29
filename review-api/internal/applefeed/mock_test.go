package applefeed

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

// MockServer can be used to mock Apple feed API responses for testing purposes.
// It spawns a local HTTP server that serves mock responses from JSON files.
type MockServer struct {
	server *httptest.Server
}

func NewMockServer() *MockServer {
	ms := &MockServer{}
	ms.server = httptest.NewServer(http.HandlerFunc(ms.handleFunc))
	return ms
}

func (ms *MockServer) handleFunc(w http.ResponseWriter, r *http.Request) {
	filename := fmt.Sprintf("data/page%s.json", ms.getPage(r))
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, "File not found: "+filename, http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/json")
	_, _ = io.Copy(w, file)
}

func (ms *MockServer) getPage(r *http.Request) string {
	// Extract page number from URL path
	// Path example: /us/rss/customerreviews/id=595068606/sortBy=mostRecent/page=1/json
	page := "1"
	start := strings.Index(r.URL.Path, "/page=")
	if start != -1 {
		start += 6
		end := strings.Index(r.URL.Path[start:], "/")
		if end != -1 {
			page = r.URL.Path[start : start+end]
		}
	}
	return page
}

func (ms *MockServer) URL() string {
	return ms.server.URL
}

func (ms *MockServer) Close() {
	ms.server.Close()
}
