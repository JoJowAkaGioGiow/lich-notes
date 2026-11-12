package main

import (
	"testing"

	"github.com/lichdotdev/lich/extension"
	"github.com/lichdotdev/lich/extension/i18n"
)

// openApp is an app on a vault holding one note, opened and typed into.
func openApp(t *testing.T, m mode) *app {
	t.Helper()
	cat, err := i18n.OpenDir(".")
	if err != nil {
		t.Fatal(err)
	}
	vault := t.TempDir()
	path, err := CreateNote(vault)
	if err != nil {
		t.Fatal(err)
	}
	a := &app{cat: cat, lang: i18n.FromEnv(), vault: vault, mode: m, split: defaultSplit,
		send: func(extension.Message) {}}
	a.open(path)
	a.mode = m
	a.text = "# typed"
	return a
}

func TestSavingAnEditShowsThePreview(t *testing.T) {
	a := openApp(t, modeEdit)
	a.handle(extension.FireEvent.New(extension.EventData{Handler: onSave}))
	if a.saved != "# typed" || a.mode != modePreview {
		t.Errorf("after save: saved %q, mode %q; want the text saved and the preview", a.saved, a.mode)
	}
}

func TestSavingASplitKeepsBothSides(t *testing.T) {
	a := openApp(t, modeSplit)
	a.handle(extension.FireEvent.New(extension.EventData{Handler: onSave}))
	if a.saved != "# typed" || a.mode != modeSplit {
		t.Errorf("after save: saved %q, mode %q; want the text saved and the split kept", a.saved, a.mode)
	}
}

func TestAnUnknownModeChangesNothing(t *testing.T) {
	a := openApp(t, modeSplit)
	a.handle(extension.FireEvent.New(extension.EventData{Handler: onMode, Text: "sideways"}))
	if a.mode != modeSplit {
		t.Errorf("mode %q, want split kept", a.mode)
	}
}
