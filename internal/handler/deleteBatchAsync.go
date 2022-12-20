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
	h.logger.Debug(string(b))
	if len(b) > 0 {
		var shorts []string
		if err = json.Unmarshal(b, &shorts); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		_ = cookie.Value
		//h.Storage.DeleteURLs(r.Context(), user, shorts)
		w.WriteHeader(http.StatusAccepted)
	}
	w.WriteHeader(http.StatusBadRequest)
}
