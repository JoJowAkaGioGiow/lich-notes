// Command lich-notes is a lich extension: a screen that lists, shows and
// edits the markdown notes of a folder.
package main

import (
	"log"
	"os"

	"github.com/lichdotdev/lich/extension"
	"github.com/lichdotdev/lich/extension/i18n"
)

func main() {
	// lich starts an extension in its own directory, where locales/ is.
	cat, err := i18n.OpenDir(".")
	if err != nil {
		log.Fatal(err)
	}
	dataDir := os.Getenv(extension.DataDirEnv)
	settings, err := LoadSettings(dataDir)
	if err != nil {
		log.Printf("settings: %v", err)
	}

	conn := extension.NewConn(os.Stdin, os.Stdout)
	a := &app{
		cat: cat, lang: i18n.FromEnv(), dataDir: dataDir,
		vault: settings.Vault, picking: settings.Vault == "",
		send: func(msg extension.Message) {
			if err := conn.Send(msg); err != nil {
				log.Printf("send: %v", err)
			}
		},
	}
	// An extension restarted under an open screen gets no surface.opened,
	// only the language, so the notes are listed before the first render.
	a.reload()
	for msg := range conn.Messages() {
		a.handle(msg)
	}
	// lich closed stdin: it is stopping the extension, and an unsaved note
	// would go with it.
	a.save()
}
