package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_StringOps(t *testing.T) {
	// mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/string/foo":
			// verify request body contains JSON with correct value
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"value":"bar"`) {
				t.Errorf("expected value bar, got %s", body)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/string/foo":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":"bar"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/string/foo":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected call: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	cli, err := NewClient(ts.URL, "my-secret-token")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// SetString
	if err := cli.SetString(ctx, "foo", "bar", 0); err != nil {
		t.Fatalf("SetString: %v", err)
	}

	// GetString
	val, err := cli.GetString(ctx, "foo")
	if err != nil {
		t.Fatalf("GetString: %v", err)
	}
	if val != "bar" {
		t.Errorf("GetString returned %q, want bar", val)
	}

	// DeleteString
	if err := cli.DeleteString(ctx, "foo"); err != nil {
		t.Fatalf("DeleteString: %v", err)
	}
}

func TestClient_ListOps(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/list/mylist/push":
			// check payload
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"items":["one","two"]`) {
				t.Errorf("unexpected body: %s", body)
			}
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/list/mylist/pop":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"value":"two"}`))
		default:
			t.Fatalf("unexpected call: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	cli, _ := NewClient(ts.URL, "my-secret-token")
	ctx := context.Background()

	if err := cli.LPush(ctx, "mylist", "one", "two"); err != nil {
		t.Fatalf("LPush: %v", err)
	}
	v, err := cli.RPop(ctx, "mylist")
	if err != nil {
		t.Fatalf("RPop: %v", err)
	}
	if v != "two" {
		t.Errorf("RPop returned %q, want two", v)
	}
}
