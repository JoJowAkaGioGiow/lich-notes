package main

import (
	"os"
	"testing"

	"github.com/lichdotdev/lich/extension/i18n"
)

func TestEveryLanguageHasEveryMessage(t *testing.T) {
	fsys := os.DirFS(".")
	source, err := i18n.Load(fsys, "locales/en")
	if err != nil {
		t.Fatal(err)
	}
	if len(source) == 0 {
		t.Fatal("locales/en is empty")
	}
	translated, err := i18n.Load(fsys, "locales/pt-br")
	if err != nil {
		t.Fatal(err)
	}
	for key := range source {
		if _, ok := translated[key]; !ok {
			t.Errorf("pt-br lacks %q", key)
		}
	}
	for key := range translated {
		if _, ok := source[key]; !ok {
			t.Errorf("pt-br has %q, which en does not", key)
		}
	}
}
