package gzcli

import "testing"

func TestGenerateUsername_PreservesInputWhenUnique(t *testing.T) {
	existing := map[string]struct{}{}

	got, err := generateUsername("Alice_01", 15, existing)
	if err != nil {
		t.Fatalf("generateUsername returned error: %v", err)
	}
	if got != "Alice_01" {
		t.Fatalf("expected username to stay unchanged, got %q", got)
	}
}

func TestGenerateUsername_AppendsSuffixWhenDuplicate(t *testing.T) {
	existing := map[string]struct{}{
		"Alice_01": {},
	}

	got, err := generateUsername("Alice_01", 15, existing)
	if err != nil {
		t.Fatalf("generateUsername returned error: %v", err)
	}
	if got != "Alice_011" {
		t.Fatalf("expected duplicate username to get suffix, got %q", got)
	}
}

func TestGenerateUsername_EmptyAfterTrim(t *testing.T) {
	existing := map[string]struct{}{}

	if _, err := generateUsername("   ", 15, existing); err == nil {
		t.Fatal("expected error for empty username")
	}
}
