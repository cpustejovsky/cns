package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cpustejovksy/cns/logger"
	"github.com/cpustejovksy/cns/store"
)

var s store.Store = store.New()

// KeyValueGetHanlder expect to be called with a GET request for the "/v1/key/{key} resource"
func KeyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
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
	key := r.PathValue("key")
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
	key := r.PathValue("key")
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

func initializeTransactionLog(s *store.Store) error {
	ftl, err := logger.NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}
	events, errors := ftl.ReadEvents()
	e := logger.Event{}
	ok := true

	for ok && err != nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case logger.EventDelete:
				err = s.Delete(e.Key)
			case logger.EventPut:
				err = s.Put(e.Key, e.Value)
			}
		}
	}

	ftl.Run()

	return err
}
func main() {
	err := initializeTransactionLog(&s)
	if err != nil {
		log.Fatal(err)
	}
	r := http.NewServeMux()
	r.HandleFunc(http.MethodPut+" "+"/v1/key/{key}", KeyValuePutHandler)
	r.HandleFunc(http.MethodGet+" "+"/v1/key/{key}", KeyValueGetHandler)
	r.HandleFunc(http.MethodDelete+" "+"/v1/key/{key}", KeyValueDeleteHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
