package ui

import (
	"fmt"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// defaultTokenTTL pre-fills the "expires in" field of the token form.
const defaultTokenTTL = time.Hour

// tokenFormPage prompts for a token's grants and expiry, then mints it via
// CreateToken and shows the result. Minting is a local JWT signing
// operation using the context's configured API key/secret (no LiveKit API
// call), so unlike the other participant actions it's available even when
// the context is read-only.
func tokenFormPage(n nav, identity, room string) tview.Primitive {
	const pageName = "token-form"

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(fmt.Sprintf(" generate token: %s @ %s ", identity, room))

	cancel := func() { n.pages.RemovePage(pageName) }

	form.AddInputField("expires in (e.g. 1h, 30m)", defaultTokenTTL.String(), 20, nil, nil)
	form.AddCheckbox("CanPublish", true, nil)
	form.AddCheckbox("CanSubscribe", true, nil)
	form.AddCheckbox("CanPublishData", true, nil)
	form.AddCheckbox("CanUpdateOwnMetadata", false, nil)
	form.AddCheckbox("Hidden", false, nil)
	form.AddCheckbox("Recorder", false, nil)

	form.AddTextView("", "", 40, 1, true, false)
	statusIdx := form.GetFormItemCount() - 1

	setStatus := func(msg string) {
		if tv, ok := form.GetFormItem(statusIdx).(*tview.TextView); ok {
			tv.SetText(msg)
		}
	}

	checkbox := func(label string) bool {
		return form.GetFormItemByLabel(label).(*tview.Checkbox).IsChecked()
	}

	generate := func() {
		ttlText := form.GetFormItemByLabel("expires in (e.g. 1h, 30m)").(*tview.InputField).GetText()

		ttl, err := time.ParseDuration(ttlText)
		if err != nil || ttl <= 0 {
			setStatus("[red] invalid duration (e.g. 1h, 30m, 24h)")

			return
		}

		grant := lk.TokenGrant{
			CanPublish:           checkbox("CanPublish"),
			CanSubscribe:         checkbox("CanSubscribe"),
			CanPublishData:       checkbox("CanPublishData"),
			CanUpdateOwnMetadata: checkbox("CanUpdateOwnMetadata"),
			Hidden:               checkbox("Hidden"),
			Recorder:             checkbox("Recorder"),
		}

		token, err := n.client.CreateToken(identity, room, ttl, grant)
		if err != nil {
			setStatus("[red] " + err.Error())

			return
		}

		n.pages.RemovePage(pageName)
		n.pages.AddPage("token", tokenResultPage(n, identity, room, ttl, token), true, true)
	}

	form.AddButton("generate", generate)
	form.AddButton("cancel", cancel)
	form.SetCancelFunc(cancel)

	return centered(form, 18, 60)
}

// tokenResultPage shows a minted token in a plain, selectable text view so
// it can be copied via the terminal's own text selection.
func tokenResultPage(n nav, identity, room string, ttl time.Duration, token string) tview.Primitive {
	const pageName = "token"

	body := tview.NewTextView().SetText(token).SetDynamicColors(true).SetScrollable(true)
	body.SetBorder(true).SetTitle(fmt.Sprintf(" token: %s @ %s (valid %s) ", identity, room, ttl))

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
