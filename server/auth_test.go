package server

import (
	"data_storage/server/adapters/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

// dummyHandler simply returns 200 OK.
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestTokenAuthMiddleware(t *testing.T) {
	expectedToken := "secret-token"
	wrapped := middleware.TokenAuth(expectedToken)(http.HandlerFunc(dummyHandler))

	t.Run("No Authorization header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}

		// WWW-Authenticate header
		authHdr := resp.Header.Get("WWW-Authenticate")
		want := `Bearer realm="restricted"`
		if authHdr != want {
			t.Errorf("expected WWW-Authenticate %q, got %q", want, authHdr)
		}
	})

	t.Run("Wrong token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/protected", nil)
		r.Header.Set("Authorization", "Bearer wrong-token")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401 for wrong token, got %d", w.Result().StatusCode)
		}
	})

	t.Run("Correct token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/protected", nil)
		r.Header.Set("Authorization", "Bearer "+expectedToken)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200 for correct token, got %d", resp.StatusCode)
		}

		body := w.Body.String()
		if body != "ok" {
			t.Errorf("expected body 'ok', got %q", body)
		}
	})
}
