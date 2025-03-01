package requestutil

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

// SingleHostReverseProxy will insert an X-Forwarded-For header, and can be used to test
// RemoteAddr().  A fake RemoteAddr cannot be set on the HTTP request - it is overwritten
// at the transport layer to 127.0.0.1:<port> .  However, as the X-Forwarded-For header
// just contains the IP address, it is different enough for testing.
func TestRemoteAddr(t *testing.T) {
	var expectedRemote string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.RemoteAddr == expectedRemote {
			t.Errorf("Unexpected matching remote addresses")
		}

		actualRemote := RemoteAddr(r)
		if expectedRemote != actualRemote {
			t.Errorf("Mismatching remote hosts: %v != %v", expectedRemote, actualRemote)
		}

		w.WriteHeader(200)
	}))

	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	frontend := httptest.NewServer(proxy)
	defer frontend.Close()

	// X-Forwarded-For set by proxy
	expectedRemote = "127.0.0.1"
	proxyReq, err := http.NewRequest(http.MethodGet, frontend.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// RemoteAddr in X-Real-Ip
	getReq, err := http.NewRequest(http.MethodGet, backend.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedRemote = "1.2.3.4"
	getReq.Header["X-Real-ip"] = []string{expectedRemote}
	resp, err = http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Valid X-Real-Ip and invalid X-Forwarded-For
	getReq.Header["X-forwarded-for"] = []string{"1.2.3"}
	resp, err = http.DefaultClient.Do(getReq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
}
