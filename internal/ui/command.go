package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// commands are the recognized ":" command names, offered as autocomplete
// entries in the order shown.
var commands = []string{"create-room", "projects", "rooms", "quit"}

// commandBarPage shows a k9s-style ":" command line with autocompletion. It
// closes itself on Esc or an empty submit; onExec is called with the
// trimmed, lower-cased command once the user submits non-empty text.
func commandBarPage(n nav, onExec func(cmd string)) tview.Primitive {
	const pageName = "command"

	input := tview.NewInputField().SetLabel(" : ")

	input.SetAutocompleteFunc(func(currentText string) []string {
		currentText = strings.ToLower(strings.TrimSpace(currentText))
		if currentText == "" {
			return nil
		}

		var matches []string

		for _, c := range commands {
			if strings.HasPrefix(c, currentText) {
				matches = append(matches, c)
			}
		}

		return matches
	})

	submit := func(text string) {
		n.pages.RemovePage(pageName)

		text = strings.ToLower(strings.TrimSpace(text))
		if text == "" {
			return
		}

		onExec(text)
	}

	// Enter on a highlighted autocomplete entry both fills and submits;
	// arrow-key navigation only fills, keeping the list open.
	input.SetAutocompletedFunc(func(text string, index int, source int) bool {
		input.SetText(text)

		if source == tview.AutocompletedNavigate {
			return false
		}

		if source == tview.AutocompletedEnter {
			submit(text)
		}

		return true
	})

	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			n.pages.RemovePage(pageName)
		case tcell.KeyEnter:
			submit(input.GetText())
		}
	})

	body := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(input, 1, 0, true)
	body.SetBorder(true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(body, 3, 0, true).
		AddItem(nil, 0, 1, false)
}
