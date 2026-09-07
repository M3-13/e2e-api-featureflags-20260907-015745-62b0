package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestAddGet(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "a", Enabled: true, Description: "desc", RolloutPercent: 50}
	if err := s.Add(f); err != nil {
		t.Fatalf("Add: %v", err)
	}
	got, ok := s.Get("a")
	if !ok {
		t.Fatal("Get: key not found")
	}
	if got != f {
		t.Fatalf("got %+v, want %+v", got, f)
	}
}

func TestAddDuplicate(t *testing.T) {
	s := NewStore()
	if err := s.Add(Flag{Key: "a"}); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if err := s.Add(Flag{Key: "a"}); err != ErrFlagExists {
		t.Fatalf("second Add err = %v, want ErrFlagExists", err)
	}
}

func TestListSorted(t *testing.T) {
	s := NewStore()
	for _, k := range []string{"c", "a", "b"} {
		if err := s.Add(Flag{Key: k}); err != nil {
			t.Fatalf("Add(%q): %v", k, err)
		}
	}
	got := s.List()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i, want := range []string{"a", "b", "c"} {
		if got[i].Key != want {
			t.Fatalf("got[%d].Key = %q, want %q", i, got[i].Key, want)
		}
	}
}

func TestUpdate(t *testing.T) {
	s := NewStore()
	if err := s.Add(Flag{Key: "a", Enabled: false, Description: "old", RolloutPercent: 10}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	f, ok := s.Update("a", true, "new", 80)
	if !ok {
		t.Fatal("Update: key not found")
	}
	if !f.Enabled || f.Description != "new" || f.RolloutPercent != 80 {
		t.Fatalf("got %+v", f)
	}
	if _, ok := s.Update("missing", true, "", 0); ok {
		t.Fatal("Update missing key: want ok == false")
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()
	if err := s.Add(Flag{Key: "a"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !s.Delete("a") {
		t.Fatal("Delete existing: want true")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("Get after Delete: want not found")
	}
	if s.Delete("missing") {
		t.Fatal("Delete missing: want false")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewStore()
	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("flag-%d", i)
			_ = s.Add(Flag{Key: key, Enabled: true})
			if _, ok := s.Get(key); !ok {
				t.Errorf("Get(%q) not found", key)
			}
			if _, ok := s.Update(key, false, "updated", 75); !ok {
				t.Errorf("Update(%q) not found", key)
			}
			s.List()
		}(i)
	}
	wg.Wait()
}
