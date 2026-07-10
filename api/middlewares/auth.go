package middlewares

import "net/http"

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value("is_admin").(bool)
		if !ok || !isAdmin {
			http.Error(w, "Forbidden access", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
