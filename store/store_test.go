package store_test

import (
	"errors"
	"testing"

	"github.com/cpustejovksy/cns/store"
)

func TestStore(t *testing.T) {
	var s store.Store
	t.Run("New", func(t *testing.T) {
		s = store.New()
	})
	t.Run("Put", func(t *testing.T) {
		err := s.Put("foo", "bar")
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Get", func(t *testing.T) {
		key := "foo"
		val, err := s.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		if got, expect := val, "bar"; got != expect {
			t.Fatalf("got %v, expect %v", got, expect)
		}
	})
	t.Run("Get Nonexistent Key", func(t *testing.T) {
		key := "food"
		val, err := s.Get(key)
		if err == nil {
			t.Fatal(err)
		}
		var compareErr store.ErrorNoSuchKey
		if !errors.As(err, &compareErr) {
			t.Fatalf("got %T, expect %T", err, compareErr)
		}
		if got, expect := val, ""; got != expect {
			t.Fatalf("got %v, expect %v", got, expect)
		}
	})
	t.Run("Delete", func(t *testing.T) {
		err := s.Delete("foo")
		if err != nil {
			t.Fatal(err)
		}
	})
}
