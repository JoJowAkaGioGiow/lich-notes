package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func writeNote(t *testing.T, vault, rel, text string, age time.Duration) {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

func paths(notes []Note) []string {
	out := make([]string, len(notes))
	for i, n := range notes {
		out[i] = n.Path
	}
	return out
}

func TestListNotesNewestFirstSkippingDotDirsAndOtherFiles(t *testing.T) {
	vault := t.TempDir()
	writeNote(t, vault, "old.md", "", 3*time.Hour)
	writeNote(t, vault, "projects/new.md", "", time.Hour)
	writeNote(t, vault, "middle.MD", "", 2*time.Hour)
	writeNote(t, vault, ".editor/workspace.md", "", 0)
	writeNote(t, vault, "projects/.trash/gone.md", "", 0)
	writeNote(t, vault, "image.png", "", 0)

	notes, err := ListNotes(vault)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"projects/new.md", "middle.MD", "old.md"}
	if got := paths(notes); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if notes[0].Title != "new" || notes[1].Title != "middle" {
		t.Fatalf("titles %q and %q, want the file names without the extension", notes[0].Title, notes[1].Title)
	}
}

func TestListNotesOfAMissingVaultFails(t *testing.T) {
	if _, err := ListNotes(filepath.Join(t.TempDir(), "gone")); err == nil {
		t.Fatal("want an error for a vault that does not exist")
	}
}

func TestSearchNotesMatchesTitleOrContentIgnoringCase(t *testing.T) {
	vault := t.TempDir()
	writeNote(t, vault, "Groceries.md", "milk and eggs", time.Hour)
	writeNote(t, vault, "Reading.md", "a paper on GROCERY pricing", 2*time.Hour)
	writeNote(t, vault, "Ideas.md", "nothing here", 3*time.Hour)
	notes, err := ListNotes(vault)
	if err != nil {
		t.Fatal(err)
	}

	for query, want := range map[string][]string{
		"grocer":  {"Groceries.md", "Reading.md"},
		"  EGGS ": {"Groceries.md"},
		"absent":  {},
		"":        {"Groceries.md", "Reading.md", "Ideas.md"},
	} {
		if got := paths(SearchNotes(vault, notes, query)); !slices.Equal(got, want) {
			t.Errorf("query %q: got %v, want %v", query, got, want)
		}
	}
}

func TestSaveNoteWritesWhatReadNoteReads(t *testing.T) {
	vault := t.TempDir()
	writeNote(t, vault, "sub/note.md", "before", 0)

	if err := SaveNote(vault, "sub/note.md", "# after\n"); err != nil {
		t.Fatal(err)
	}
	got, err := ReadNote(vault, "sub/note.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != "# after\n" {
		t.Fatalf("got %q", got)
	}
}

func TestNotesOutsideTheVaultAreRefused(t *testing.T) {
	root := t.TempDir()
	vault := filepath.Join(root, "vault")
	writeNote(t, root, "secret.md", "secret", 0)
	writeNote(t, vault, "note.md", "", 0)

	for _, rel := range []string{"../secret.md", filepath.Join(root, "secret.md"), ""} {
		if _, err := ReadNote(vault, rel); !errors.Is(err, errOutsideVault) {
			t.Errorf("read %q: got %v, want errOutsideVault", rel, err)
		}
		if err := SaveNote(vault, rel, "overwritten"); !errors.Is(err, errOutsideVault) {
			t.Errorf("save %q: got %v, want errOutsideVault", rel, err)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(root, "secret.md")); string(got) != "secret" {
		t.Fatalf("the file outside the vault now holds %q", got)
	}
}

func TestCreateNoteTakesTheFirstFreeNumber(t *testing.T) {
	vault := t.TempDir()
	writeNote(t, vault, "Untitled 2.md", "kept", 0)

	for _, want := range []string{"Untitled 1.md", "Untitled 3.md", "Untitled 4.md"} {
		got, err := CreateNote(vault)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
	if text, _ := ReadNote(vault, "Untitled 2.md"); text != "kept" {
		t.Fatalf("the existing note now holds %q", text)
	}
}
