package discoverd

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPClient(t *testing.T) {
	client, cleanup := setup(t)
	defer cleanup()

	hc := client.HTTPClient()
	_, err := hc.Get("http://httpclient/")
	if ue, ok := err.(*url.Error); !ok || ue.Err != ErrNoServices {
		t.Error("Expected err to be ErrNoServices, got", ue.Err)
	}

	s := httptest.NewServer(nil)
	defer s.Close()
	client.Register("httpclient", s.URL[7:])

	_, err = hc.Get("http://httpclient/")
	if err != nil {
		t.Error("Unexpected error during request:", err)
	}
}