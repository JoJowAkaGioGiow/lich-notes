package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsOfAFirstRunAreEmpty(t *testing.T) {
	got, err := LoadSettings(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != (Settings{}) {
		t.Fatalf("got %+v", got)
	}
}

func TestSettingsSurviveARestart(t *testing.T) {
	dir := t.TempDir()
	want := Settings{Vault: "/home/someone/notes"}
	if err := SaveSettings(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestSettingsThatAreNotJSONFail(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, settingsFile), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSettings(dir); err == nil {
		t.Fatal("want an error for a broken settings file")
	}
}

func TestResolveVault(t *testing.T) {
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	vault := filepath.Join(home, "notes")
	if err := os.Mkdir(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(home, "file.md")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	for typed, want := range map[string]string{vault: vault, "  " + vault + " ": vault, "~/notes": vault, "~": home} {
		got, err := ResolveVault(typed)
		if err != nil {
			t.Errorf("%q: %v", typed, err)
			continue
		}
		if got != want {
			t.Errorf("%q: got %q, want %q", typed, got, want)
		}
	}
	link := filepath.Join(home, "link")
	if err := os.Symlink(vault, link); err != nil {
		t.Fatal(err)
	}
	if got, err := ResolveVault(link); err != nil || got != vault {
		t.Errorf("a link to the vault: got %q, %v, want %q", got, err, vault)
	}
	if _, err := ResolveVault(file); !errors.Is(err, errNotFolder) {
		t.Errorf("a file: got %v, want errNotFolder", err)
	}
	if _, err := ResolveVault(filepath.Join(home, "gone")); err == nil {
		t.Error("want an error for a path that does not exist")
	}
}
