package logger_test

import (
	"os"
	"testing"

	"github.com/cpustejovksy/cns/logger"
)

const newLogger = "/tmp/new_logger.log"
const writeAppend = "/tmp/write_append.log"
const writePut = "/tmp/write_put.log"
const key = "my-key"

// Initially I am expect no events and no errors for ReadEvents()
// AFter running
func TestFileTransactionLogger(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Cleanup(func() {
			os.Remove(newLogger)
		})
		ftl, err := logger.NewFileTransactionLogger(newLogger)
		if err != nil {
			t.Fatal(err)
		}
		if ftl == nil {
			var compare *logger.FileTransactionLogger
			t.Fatalf("got %T, expected %T", ftl, compare)
		}
	})
	t.Run("Handles white space in keys and values", func(t *testing.T) {
		key := "my spacey key"
		value := "newline\tvalue\n"
		t.Cleanup(func() {
			os.Remove(newLogger)
		})
		ftl, err := logger.NewFileTransactionLogger(newLogger)
		if err != nil {
			t.Fatal(err)
		}
		ftl.Run()
		defer ftl.Close()
		chev, cherr := ftl.ReadEvents()
		for e := range chev {
			if got, expect := e.Key, key; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
		}
		err = <-cherr
		if err != nil {
			t.Error(err)
		}
		ftl.WritePut(key, value)
		ftl.WritePut(key, value)
		ftl.Wait()
		ftl2, err := logger.NewFileTransactionLogger(writeAppend)
		if err != nil {
			t.Fatal(err)
		}
		ftl2.Run()
		defer ftl2.Close()
		chev, cherr = ftl2.ReadEvents()
		for e := range chev {
			if got, expect := e.Key, key; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
			if got, expect := e.Value, value; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
		}
		err = <-cherr
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Appending Writes", func(t *testing.T) {
		t.Cleanup(func() {
			os.Remove(writeAppend)
		})
		ftl, err := logger.NewFileTransactionLogger(writeAppend)
		if err != nil {
			t.Fatal(err)
		}
		ftl.Run()
		defer ftl.Close()
		chev, cherr := ftl.ReadEvents()
		for e := range chev {
			if got, expect := e.Key, key; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
		}
		err = <-cherr
		if err != nil {
			t.Error(err)
		}
		ftl.WritePut(key, "my-value")
		ftl.WritePut(key, "my-value2")
		ftl.Wait()
		ftl2, err := logger.NewFileTransactionLogger(writeAppend)
		if err != nil {
			t.Fatal(err)
		}
		ftl2.Run()
		defer ftl2.Close()
		chev, cherr = ftl2.ReadEvents()
		for e := range chev {
			if got, expect := e.Key, key; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
		}
		err = <-cherr
		if err != nil {
			t.Error(err)
		}
		ftl2.WritePut(key, "my-value3")
		ftl2.WritePut(key, "my-value4")
		ftl2.Wait()

		if ftl2.LastSequence() != 4 {
			t.Errorf("Last sequence mismatch (expected 4; got %d)", ftl.LastSequence())
		}
	})
	t.Run("AppendPut", func(t *testing.T) {
		t.Cleanup(func() {
			os.Remove(writePut)
		})
		ftl, err := logger.NewFileTransactionLogger(writePut)
		if err != nil {
			t.Fatal(err)
		}
		ftl.Run()
		defer ftl.Close()
		chev, cherr := ftl.ReadEvents()
		for e := range chev {
			if got, expect := e.Key, key; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
		}
		err = <-cherr
		if err != nil {
			t.Error(err)
		}
		ftl.WritePut(key, "my-value")
		ftl.WritePut(key, "my-value2")
		ftl.WritePut(key, "my-value3")
		ftl.WritePut(key, "my-value4")
		ftl.Wait()
		ftl2, err := logger.NewFileTransactionLogger(writePut)
		if err != nil {
			t.Fatal(err)
		}
		defer ftl2.Close()
		chev, cherr = ftl2.ReadEvents()
		for e := range chev {
			if got, expect := e.Key, key; got != expect {
				t.Fatalf("got '%s', expected '%s'", got, expect)
			}
		}
		err = <-cherr
		if err != nil {
			t.Error(err)
		}
		ftl.Wait()

		if got, expect := ftl.LastSequence(), ftl2.LastSequence(); got != expect {
			t.Errorf("Last sequence mismatch (got %d, expectected %d)", got, expect)
		}
	})

	// s := store.New()
	// 	events, errors := ftl.ReadEvents()
	// 	ok := true
	//
	// 	for ok && err != nil {
	// 		select {
	// 		case err, ok = <-errors:
	// 			t.Fatal(err)
	// 		case _, ok = <-events:
	// 			t.Fatal("expected no events on an empty file")
	// 		}
	// 	}
	// 	ftl.Run()
	// 	ftl.WritePut("foo", "bar")
	// 	ftl.WriteDelete("foo")
	// 	events, errors = ftl.ReadEvents()
	// 	var count int
	// 	for count < 2 {
	// 		select {
	// 		case err, ok = <-errors:
	// 			if err != nil && ok {
	// 				t.Fatal(err)
	// 			}
	// 		case e, ok := <-events:
	// 			if !ok {
	// 				t.Fatal("loop should break before channel is drained")
	// 			}
	// 			switch e.EventType {
	// 			case logger.EventDelete:
	// 				count++
	// 			case logger.EventPut:
	// 				count++
	// 			}
	// 		}
	// 	}
}
