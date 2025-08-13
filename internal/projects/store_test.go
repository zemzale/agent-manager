package projects

import (
	"path/filepath"
	"testing"
)

func TestStore_AddGetListRemove(t *testing.T) {
	base := t.TempDir()
	s, err := NewStoreWithBaseDir(base)
	if err != nil {
		t.Fatalf("NewStoreWithBaseDir: %v", err)
	}

	// Initially empty
	items, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %d", len(items))
	}

	// Add
	if err := s.Add("myapp", "git@github.com:user/myapp.git"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Add duplicate should fail
	if err := s.Add("myapp", "git@github.com:user/myapp.git"); err == nil {
		t.Fatalf("expected duplicate add to fail")
	}

	// Get
	u, ok, err := s.Get("myapp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok || u != "git@github.com:user/myapp.git" {
		t.Fatalf("Get returned %v, %v", u, ok)
	}

	// Set overwrite
	if err := s.Set("myapp", "https://github.com/user/myapp.git"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	u, ok, err = s.Get("myapp")
	if err != nil {
		t.Fatalf("Get after Set: %v", err)
	}
	if !ok || u != "https://github.com/user/myapp.git" {
		t.Fatalf("Get after Set returned %v, %v", u, ok)
	}

	// List has one
	items, err = s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 || items[0].Name != "myapp" {
		t.Fatalf("List unexpected: %#v", items)
	}

	// Remove
	if err := s.Remove("myapp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok, err = s.Get("myapp")
	if err != nil {
		t.Fatalf("Get after Remove: %v", err)
	}
	if ok {
		t.Fatalf("expected removed project to be absent")
	}

	// Ensure file was created in base
	if _, err := filepath.Abs(filepath.Join(base, "config.json")); err != nil {
		t.Fatalf("abs path: %v", err)
	}
}
