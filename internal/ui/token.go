package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// tokenTTL is the validity period of tokens minted from the "generate
// token" action; it's a debugging aid, not a token minted for production
// use, so a short, fixed TTL keeps that clear.
const tokenTTL = time.Hour

// tokenPage mints and displays a room-join access token for identity in
// room. This is a local JWT signing operation only (no LiveKit API call),
// so unlike the other participant actions it's available even when the
// context is read-only. The token is shown in a plain, selectable text view
// so it can be copied via the terminal's own text selection.
func tokenPage(n nav, identity, room string) tview.Primitive {
	const pageName = "token"

	token, err := n.client.CreateToken(identity, room, tokenTTL)

	content := token
	if err != nil {
		content = "[red] " + err.Error()
	}

	body := tview.NewTextView().SetText(content).SetDynamicColors(true).SetScrollable(true)
	body.SetBorder(true).SetTitle(fmt.Sprintf(" token: %s @ %s (valid %s) ", identity, room, tokenTTL))

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage(pageName)

			return nil
		}

		return event
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}
