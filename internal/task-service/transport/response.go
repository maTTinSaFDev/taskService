package transport

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode JSON response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
