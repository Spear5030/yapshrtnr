package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

func (h *Handler) DeleteBatchByUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var shorts []string
	if err := json.Unmarshal(b, &shorts); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	user := cookie.Value
	h.Storage.DeleteURLs(r.Context(), user, shorts)
	w.WriteHeader(http.StatusAccepted)
}
