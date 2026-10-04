package components

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/rivo/tview"
)

type KeyValueView struct {
	View *tview.TextView
}

func NewKeyValueView() *KeyValueView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.ColorBg)
	tv.SetBorderPadding(0, 0, 2, 1)
	return &KeyValueView{View: tv}
}

func (kv *KeyValueView) Update(info *models.KeyValueInfo) {
	if info == nil {
		kv.View.SetText("[#7d8590]No key selected[-]")
		return
	}

	header := fmt.Sprintf(
		"[#7d8590]Key:[-]  [#c9d1d9]%s[-]\n"+
			"[#7d8590]Type:[-] [#8b949e]%s[-]\n"+
			"[#7d8590]Size:[-] [#f0883e]%s[-]\n"+
			"[#7d8590]TTL:[-]  [#3fb950]%s[-]\n",
		tview.Escape(info.Key),
		info.Type,
		utils.FormatBytes(info.Size),
		formatTTL(info.TTL),
	)

	if info.Length > 0 {
		header += fmt.Sprintf("[#7d8590]Len:[-]  [#8b949e]%d[-]\n", info.Length)
	}

	header += "\n[#7d8590]─── Value ───[-]\n\n"
	header += fmt.Sprintf("[#c9d1d9]%s[-]", tview.Escape(info.Value))

	kv.View.SetText(header)
	kv.View.ScrollToBeginning()
}

func formatTTL(ttl int64) string {
	if ttl < 0 {
		return "no expiry"
	}
	return utils.FormatDuration(ttl)
}
