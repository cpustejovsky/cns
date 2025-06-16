package client

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type User struct {
	First, Last string
}

func HandlePostUser() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(r io.ReadCloser) {
			_, _ = io.Copy(io.Discard, r)
			_ = r.Close()
		}(r.Body)

		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			log.Println(err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
