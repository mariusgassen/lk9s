package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/beelis/lk9s/internal/lk"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func tracksPage(n nav, identity string, tracks []lk.Track) tview.Primitive {
	var b strings.Builder

	if len(tracks) == 0 {
		b.WriteString("(no published tracks)\n")
	}

	for i, t := range tracks {
		if i > 0 {
			b.WriteString("\n")
		}

		writeTrack(&b, t)
	}

	body := tview.NewTextView().SetText(b.String()).SetScrollable(true)
	body.SetBorder(true).SetTitle(fmt.Sprintf(" published tracks: %s (%d) ", identity, len(tracks)))

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			n.pages.RemovePage("tracks")

			return nil
		}

		return event
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(body, 0, 6, true).
			AddItem(nil, 0, 1, false), 0, 4, true).
		AddItem(nil, 0, 1, false)
}

func writeTrack(b *strings.Builder, t lk.Track) {
	field := func(label string, v any) {
		fmt.Fprintf(b, "  %-18s %v\n", label, v)
	}

	fmt.Fprintf(b, "%s  %s / %s\n", t.SID, t.Type, t.Source)
	field("Name", t.Name)
	field("MimeType", t.MimeType)
	field("Muted", t.Muted)
	field("MID", t.MID)
	field("Stream", t.Stream)
	field("Encryption", t.Encryption)
	field("BackupCodecPolicy", t.BackupCodecPolicy)

	if t.Version != 0 {
		field("Version", time.UnixMicro(t.Version).Format(time.DateTime))
	}

	if t.Type == "VIDEO" {
		field("Resolution", fmt.Sprintf("%dx%d", t.Width, t.Height))
		field("Simulcast", t.Simulcast)
	}

	if t.Type == "AUDIO" {
		field("Stereo", t.Stereo)
		field("DisableDTX", t.DisableDTX)
		field("DisableRED", t.DisableRED)
		field("AudioFeatures", strings.Join(t.AudioFeatures, ", "))
	}

	writeLayers(b, "  ", t.Layers)

	for _, c := range t.Codecs {
		fmt.Fprintf(b, "  Codec %s\n", c.MimeType)
		fmt.Fprintf(b, "    %-16s %s\n", "MID", c.MID)
		fmt.Fprintf(b, "    %-16s %s\n", "CID", c.CID)

		if c.SDPCID != "" {
			fmt.Fprintf(b, "    %-16s %s\n", "SDP CID", c.SDPCID)
		}

		if t.Type == "VIDEO" {
			fmt.Fprintf(b, "    %-16s %s\n", "LayerMode", c.LayerMode)
		}

		writeLayers(b, "    ", c.Layers)
	}
}

func writeLayers(b *strings.Builder, indent string, layers []lk.VideoLayer) {
	for _, l := range layers {
		fmt.Fprintf(b, "%sLayer %-6s %4dx%-4d %7.0f kbps  ssrc=%d rtx=%d spatial=%d rid=%q\n",
			indent, l.Quality, l.Width, l.Height, float64(l.Bitrate)/1000,
			l.SSRC, l.RepairSSRC, l.SpatialLayer, l.RID)
	}
}
