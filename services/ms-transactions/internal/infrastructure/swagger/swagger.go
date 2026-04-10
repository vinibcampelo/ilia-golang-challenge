package swagger

import (
	_ "embed"
	"net/http"
)

//go:embed ui.html
var uiHTML []byte

func Register(mux *http.ServeMux, spec []byte) {
	if len(spec) == 0 {
		return
	}

	mux.Handle("/openapi.yaml", getOnly(serveSpec(spec)))
	mux.Handle("/swagger", getOnly(redirectTo("/swagger/")))
	mux.Handle("/swagger/", getOnly(serveUI()))
}

func getOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func serveSpec(spec []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(spec)
	})
}

func redirectTo(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusFound)
	})
}

func serveUI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(uiHTML)
	})
}
