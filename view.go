package main

import (
	"github.com/lichdotdev/lich/extension"
)

const (
	sidebarWidth   = 280
	vaultFormWidth = 460
	pad            = 12
	gap            = 8
	noteRowHeight  = 44
	dateLayout     = "2006-01-02"
)

type (
	node = extension.Node
	box  = extension.BoxProps
)

func (a *app) render() {
	data := extension.RenderData{Surface: surfaceID, Root: a.vaultForm()}
	if !a.picking {
		data.Root = extension.Box.New(box{Direction: extension.DirectionRow, Grow: 1},
			a.sidebar(), extension.VerticalRule.New(extension.NoProps{}), a.notePane())
		data.Shortcuts = a.shortcuts()
	}
	a.send(extension.Render.New(data))
}

func (a *app) shortcuts() []extension.Shortcut {
	keys := []extension.Shortcut{{Keys: "/", Handler: onFind}, {Keys: "ctrl+n", Handler: onNew}}
	if a.editing {
		keys = append(keys, extension.Shortcut{Keys: "ctrl+s", Handler: onSave})
	}
	return keys
}

func (a *app) vaultForm() node {
	field := extension.Input.New(extension.InputProps{
		Value: a.vaultInput, Placeholder: a.t("vault.placeholder"), Mono: true,
		Invalid: a.vaultError != "", OnChange: onVaultInput, OnSubmit: onVaultOpen,
	}).WithKey(vaultKey)
	actions := []node{extension.Button.New(extension.ButtonProps{Text: a.t("vault.open"), OnPress: onVaultOpen})}
	if a.vault != "" {
		actions = append(actions, extension.Button.New(extension.ButtonProps{
			Text: a.t("vault.cancel"), Variant: extension.VariantGhost, OnPress: onVaultCancel,
		}))
	}
	form := []node{
		extension.Label.New(extension.LabelProps{Text: a.t("vault.title"), Size: 18, Weight: extension.FontSemibold}),
		extension.Label.New(extension.LabelProps{Text: a.t("vault.description"), Color: "muted-foreground"}),
		field,
	}
	if a.vaultError != "" {
		form = append(form, extension.Label.New(extension.LabelProps{Text: a.vaultError, Color: "destructive"}))
	}
	form = append(form, extension.Box.New(box{Direction: extension.DirectionRow, Gap: gap}, actions...))
	return extension.Box.New(box{Grow: 1, Align: extension.AlignCenter, Justify: extension.JustifyCenter},
		extension.Box.New(box{Width: vaultFormWidth, Gap: pad}, form...))
}

func (a *app) sidebar() node {
	search := extension.Input.New(extension.InputProps{
		Value: a.query, Placeholder: a.t("search"), Leading: "search", Clearable: true, OnChange: onQuery,
	}).WithKey(searchKey)
	header := extension.Box.New(box{Direction: extension.DirectionRow, Gap: gap / 2, Padding: pad, Align: extension.AlignCenter},
		extension.Box.New(box{Grow: 1}, search),
		extension.GhostIconButton.New(extension.GhostIconButtonProps{Icon: "plus", Hint: a.t("new"), OnPress: onNew}),
		extension.GhostIconButton.New(extension.GhostIconButtonProps{Icon: "folder-open", Hint: a.t("vault.change"), OnPress: onVaultChange}),
	)
	return extension.Box.New(box{Width: sidebarWidth, Background: "sidebar"}, header, a.noteList())
}

func (a *app) noteList() node {
	found := SearchNotes(a.vault, a.notes, a.query)
	if len(found) == 0 {
		title := a.t("empty.vault")
		if a.query != "" {
			title = a.t("empty.search")
		}
		return extension.EmptyState.New(extension.EmptyStateProps{Icon: "notebook-text", Title: title})
	}
	rows := make([]node, len(found))
	for i, n := range found {
		rows[i] = extension.NavRow.New(
			extension.NavRowProps{On: n.Path == a.selected, Height: noteRowHeight, OnPress: extension.Handler(openPrefix + n.Path)},
			extension.TwoLines.New(extension.TwoLinesProps{Top: n.Title, Under: n.Modified.Format(dateLayout)}),
		).WithKey(n.Path)
	}
	return extension.Scroll.New(extension.ScrollProps{}, extension.Box.New(box{PaddingX: gap}, rows...))
}

func (a *app) notePane() node {
	if a.selected == "" {
		return extension.Box.New(box{Grow: 1}, extension.EmptyState.New(extension.EmptyStateProps{
			Icon: "notebook-pen", Title: a.t("empty.select"), Description: a.t("empty.selectHint"),
		}))
	}
	bar := []node{
		extension.Box.New(box{Grow: 1}, extension.Label.New(extension.LabelProps{
			Text: a.title(), Weight: extension.FontSemibold, MaxLines: 1,
		})),
		extension.Toggle.New(extension.ToggleProps{
			Pressed: a.editing, Text: a.t("edit"), Icon: "pencil", Size: extension.ToggleSizeSM, OnChange: onEdit,
		}),
		extension.Button.New(extension.ButtonProps{
			Text: a.t("save"), Size: extension.SizeSM, Disabled: a.text == a.saved, OnPress: onSave,
		}),
	}
	return extension.Box.New(box{Grow: 1},
		extension.Box.New(box{Direction: extension.DirectionRow, Gap: gap, Padding: pad, Align: extension.AlignCenter}, bar...),
		extension.Separator.New(extension.NoProps{}),
		a.noteBody(),
	)
}

// noteBody keys the field by the note, so one note's caret and undo history
// never land on another.
func (a *app) noteBody() node {
	if a.editing {
		return extension.Box.New(box{Grow: 1, Padding: pad}, extension.Textarea.New(extension.TextareaProps{
			Value: a.text, Placeholder: a.t("placeholder"), Grow: true, OnChange: onDraft,
		}).WithKey(editorKey(a.selected)))
	}
	return extension.Scroll.New(extension.ScrollProps{},
		extension.Box.New(box{Padding: pad * 2}, extension.Markdown.New(extension.MarkdownProps{Text: a.text, OnOpen: onLink})))
}
