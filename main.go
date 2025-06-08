package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cpustejovksy/cns/store"
	"github.com/gorilla/mux"
)

var s store.Store = store.New()

// KeyValueGetHanlder expect to be called with a GET request for the "/v1/key/{key} resource"
func KeyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		http.Error(w, "provide key", http.StatusBadRequest)
		return
	}
	val, err := s.Get(key)
	if err != nil {
		var compare store.ErrorNoSuchKey
		if errors.As(err, &compare) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, val)
}

// KeyValuePutHanlder expect to be called with a PUT request for the "/v1/key/{key} resource"
func KeyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		http.Error(w, "provide key", http.StatusBadRequest)
		return
	}
	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// KeyValueDeleteHandler expects to be called with a DELETE request for the "/v1/key/{key} resource"
func KeyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		http.Error(w, "provide key", http.StatusBadRequest)
		return
	}
	err := s.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/key/{key}", KeyValuePutHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/key/{key}", KeyValueGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/key/{key}", KeyValueDeleteHandler).Methods(http.MethodDelete)
	log.Fatal(http.ListenAndServe(":8080", r))
}
