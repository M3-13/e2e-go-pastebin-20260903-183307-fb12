package paste

import (
	"sync"
	"testing"
	"time"
)

func TestStoreCreateGet(t *testing.T) {
	s := NewStore()
	p, err := s.Create("hello world", "go", 0)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(p.ID) != 32 {
		t.Fatalf("expected 32-character id, got %q (%d)", p.ID, len(p.ID))
	}
	if p.Content != "hello world" {
		t.Fatalf("expected content preserved, got %q", p.Content)
	}
	if p.Language != "go" {
		t.Fatalf("expected language preserved, got %q", p.Language)
	}
	if p.ExpiresAt != nil {
		t.Fatal("expected nil ExpiresAt for non-expiring paste")
	}

	got, ok := s.Get(p.ID)
	if !ok {
		t.Fatal("expected Get to find the paste")
	}
	if got.Content != "hello world" {
		t.Fatalf("expected content hello world, got %q", got.Content)
	}
}

func TestStoreIDsUnique(t *testing.T) {
	s := NewStore()
	a, _ := s.Create("a", "", 0)
	b, _ := s.Create("b", "", 0)
	if a.ID == b.ID {
		t.Fatal("expected unique ids")
	}
}

func TestStoreExpiry(t *testing.T) {
	s := NewStore()
	p, err := s.Create("x", "", time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if p.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt set for positive expiresIn")
	}

	time.Sleep(5 * time.Millisecond)

	if _, ok := s.Get(p.ID); ok {
		t.Fatal("expected expired paste to be unavailable via Get")
	}
	if len(s.List()) != 0 {
		t.Fatal("expected expired paste to be excluded from List")
	}
	if !s.IsExpired(p) {
		t.Fatal("expected IsExpired true after expiry")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	p, _ := s.Create("x", "", 0)
	if !s.Delete(p.ID) {
		t.Fatal("expected Delete to return true for existing id")
	}
	if _, ok := s.Get(p.ID); ok {
		t.Fatal("expected Get false after Delete")
	}
	if s.Delete(p.ID) {
		t.Fatal("expected Delete to return false for unknown id")
	}
}

func TestStoreGetUnknown(t *testing.T) {
	s := NewStore()
	if _, ok := s.Get("does-not-exist"); ok {
		t.Fatal("expected Get false for unknown id")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()
	const n = 100
	ids := make([]string, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p, err := s.Create("content", "go", 0)
			if err != nil {
				t.Errorf("Create: %v", err)
				return
			}
			ids[i] = p.ID
		}(i)
	}
	wg.Wait()

	wg = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, ok := s.Get(ids[i]); !ok {
				t.Errorf("Get(%q) expected true", ids[i])
			}
			_ = s.List()
		}(i)
	}
	wg.Wait()

	wg = sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Delete(ids[i])
		}(i)
	}
	wg.Wait()
}
