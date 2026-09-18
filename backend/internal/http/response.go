package http

import (
	"encoding/json"
	"io"
	"log/slog"
	stdhttp "net/http"
)

func respond(w stdhttp.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func fail(w stdhttp.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}
func internalError(w stdhttp.ResponseWriter, err error) {
	slog.Error("internal API error", "error", err)
	fail(w, 500, "internal server error")
}
func decodeJSON(w stdhttp.ResponseWriter, r *stdhttp.Request, destination interface{}) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		fail(w, 400, "invalid JSON request body")
		return false
	}
	return true
}
