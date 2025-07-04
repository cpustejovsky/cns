package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"syscall"
	"time"

	"github.com/cpustejovksy/cns/logger"
	"github.com/cpustejovksy/cns/store"
)

var (
	dbname = flag.String("dbname", cmp.Or(os.Getenv("pg_dbname"), "test"), "database name to connect to")
	dbhost = flag.String("dbhost", cmp.Or(os.Getenv("pg_host"), host()), "database host to connect to")
	dbuser = flag.String("dbuser", cmp.Or(os.Getenv("pg_user"), me()), "database user to connect as")
	dbpw   = flag.String("dbpw", cmp.Or(os.Getenv("pg_pw"), ""), "database password")
	addr   = flag.String("addr", cmp.Or(os.Getenv("addr"), ":8080"), "address")
)

func me() string {
	me, err := user.Current()
	if err != nil {
		return ""
	}
	return me.Username
}

func host() string {
	switch runtime.GOOS {
	case "freebsd", "darwin":
		return "/tmp"
	case "linux":
		return "/var/run/postgresql"
	default:
		return ""
	}
}

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
	tl, err := logger.NewFileTransactionLogger("transaction.log")
	// param := logger.PostgresDbParams{
	// 	Host:     *dbhost,
	// 	DBName:   *dbname,
	// 	User:     *dbuser,
	// 	Password: *dbpw,
	// }
	// tl, err := logger.NewPostgresTransactionLogger(param)
	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}
	events, errors := tl.ReadEvents()
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

	tl.Run()

	return err
}
func main() {
	flag.Parse()
	err := initializeTransactionLog(&s)
	if err != nil {
		log.Fatal(err)
	}
	r := http.NewServeMux()
	r.HandleFunc(http.MethodPut+" "+"/v1/key/{key}", KeyValuePutHandler)
	r.HandleFunc(http.MethodGet+" "+"/v1/key/{key}", KeyValueGetHandler)
	r.HandleFunc(http.MethodDelete+" "+"/v1/key/{key}", KeyValueDeleteHandler)
	svr := http.Server{
		Handler: r,
		Addr:    *addr,
	}
	// run server in a goroutine so we can multiplex between signal and error
	// handling below.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("Server Started", "port", *addr)
		errCh <- svr.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case <-ctx.Done():
		slog.Error("server shutting down", "error", ctx.Err())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := svr.Shutdown(ctx)
		if err != nil {
			slog.Error("failed to shutdown server, exiting anyway", "error", err)
			os.Exit(1)

		}
		slog.Info("Server shut down successfully")

	}
}
