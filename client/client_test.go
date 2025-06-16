package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpustejovksy/cns/client"
)

func TestHeadTime(t *testing.T) {
	resp, err := http.Head("https://www.time.gov")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	now := time.Now().Round(time.Second)
	date := resp.Header.Get("Date")
	if date == "" {
		t.Fatal("no Date header received from time.gov")
	}

	dt, err := time.Parse(time.RFC1123, date)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("time.gov: %s (skew %s)\n", dt, now.Sub(dt))
}

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	/*NOTE: for a process that needs to kept alive, use:
	  ctx, cancel := context.ContextWithCancel(t.Context())
	  timer := time.AfterFunc(5*time.Second, cancel)
	  //give timer more time for X reason
	  timer.Reset(5*time.Second)
	*/
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	//NOTE: Set `req` to close to tell Go's HTTP client to close the underlying TCP conn after reading the response
	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(client.HandlePostUser()))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusMethodNotAllowed, resp.StatusCode)
	}
	_ = resp.Body.Close()

	buf := new(bytes.Buffer)
	u := client.User{First: "Adam", Last: "Woodbeck"}
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusAccepted, resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestMultipartPost(t *testing.T) {
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)

	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Attach files
	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i+1),
			filepath.Base(file))
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		_, err = io.Copy(filePart, f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	// Finalize multipart request body
	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://httpbin.org/post", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}

	t.Logf("\n%s", b)
}
