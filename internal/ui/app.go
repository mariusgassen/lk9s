package ui

import (
	"context"
	"strings"
	"syscall"
	"time"

	"os"
	"os/signal"

	"github.com/beelis/lk9s/internal/config"
	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 5 * time.Second

// RoomLister is the subset of *lk.Client the UI depends on.
type RoomLister interface {
	ListRooms(ctx context.Context) ([]lk.Room, error)
	ListParticipants(ctx context.Context, room string) ([]lk.Participant, error)
	ListEgresses(ctx context.Context, room string) ([]lk.Egress, error)
	ListAgentDispatches(ctx context.Context, room string) ([]lk.AgentDispatch, error)
	ListSIP(ctx context.Context) ([]lk.SIPEntry, error)
	DeleteRoom(ctx context.Context, room string) error
}

type nav struct {
	app         *tview.Application
	pages       *tview.Pages
	client      RoomLister
	contextName string
	version     string
	ctx         context.Context
}

// Run starts the TUI, connected to dial(current). contexts is the full list
// of configured projects, offered by the ":projects" command to switch mid
// session.
func Run(dial func(config.Context) RoomLister, contexts []config.Context, current config.Context, version string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := tview.NewApplication()
	pages := tview.NewPages()
	n := nav{app: app, pages: pages, client: dial(current), contextName: current.Name, version: version, ctx: ctx}

	// cancelRooms stops the previous rooms page's background refresh (and,
	// transitively, any subpage opened from it) when a project switch
	// replaces it, instead of leaking a poller per switch.
	var cancelRooms context.CancelFunc

	showRooms := func(nn nav) {
		if cancelRooms != nil {
			cancelRooms()
		}

		roomsCtx, roomsCancel := context.WithCancel(ctx)
		cancelRooms = roomsCancel
		nn.ctx = roomsCtx

		pages.RemovePage("rooms")
		pages.AddPage("rooms", roomsPage(nn), true, true)
		pages.SwitchToPage("rooms")
	}

	switchContext := func(c config.Context) {
		nn := n
		nn.client = dial(c)
		nn.contextName = c.Name
		showRooms(nn)
	}

	showRooms(n)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() != ':' {
			return event
		}

		if _, ok := app.GetFocus().(*tview.InputField); ok {
			return event
		}

		pages.AddPage("command", commandBarPage(n, func(cmd string) {
			switch cmd {
			case "projects":
				prev, _ := pages.GetFrontPage()

				pages.AddPage("contexts", contextsPage(n, contexts, switchContext, func() {
					pages.SwitchToPage(prev)
				}), true, true)
				pages.SwitchToPage("contexts")
			case "rooms":
				pages.SwitchToPage("rooms")
			case "quit":
				app.Stop()
			}
		}), true, true)

		return nil
	})

	return app.SetRoot(pages, true).Run()
}

func newTable(title string) *tview.Table {
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false).SetFixed(1, 0)
	table.SetTitle(title).SetBorder(true)

	return table
}

func newStatusBar() *tview.TextView {
	return tview.NewTextView().SetDynamicColors(true)
}

func updateStatus(bar *tview.TextView, err error) {
	if err != nil {
		bar.SetText("[red] error: " + err.Error())

		return
	}

	bar.SetText("")
}

func legend(entries [][2]string) *tview.TextView {
	var b strings.Builder

	for _, e := range entries {
		if b.Len() > 0 {
			b.WriteString("  ")
		}

		b.WriteString("<")
		b.WriteString(e[0])
		b.WriteString("> ")
		b.WriteString(e[1])
	}

	return tview.NewTextView().SetText(" " + b.String())
}
