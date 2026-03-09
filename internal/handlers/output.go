package handlers

import (
	"fmt"
	"strings"

	"github.com/net2share/go-corelib/tui"
	"github.com/net2share/nethopper/internal/actions"
)

// TUIOutput implements OutputWriter using the tui package.
type TUIOutput struct {
	progressView *tui.ProgressView
	interactive  bool
}

func NewTUIOutput() *TUIOutput {
	return &TUIOutput{interactive: false}
}

func NewInteractiveTUIOutput() *TUIOutput {
	return &TUIOutput{interactive: true}
}

func (t *TUIOutput) Print(msg string) {
	if t.progressView != nil {
		t.progressView.AddText(msg)
		return
	}
	fmt.Print(msg)
}

func (t *TUIOutput) Printf(format string, args ...interface{}) {
	if t.progressView != nil {
		t.progressView.AddText(fmt.Sprintf(format, args...))
		return
	}
	fmt.Printf(format, args...)
}

func (t *TUIOutput) Println(args ...interface{}) {
	if t.progressView != nil {
		if len(args) == 0 {
			t.progressView.AddText("")
		} else {
			t.progressView.AddText(fmt.Sprint(args...))
		}
		return
	}
	fmt.Println(args...)
}

func (t *TUIOutput) Info(msg string)    { t.statusMsg(msg, "info") }
func (t *TUIOutput) Success(msg string) { t.statusMsg(msg, "success") }
func (t *TUIOutput) Warning(msg string) { t.statusMsg(msg, "warning") }
func (t *TUIOutput) Error(msg string)   { t.statusMsg(msg, "error") }

func (t *TUIOutput) statusMsg(msg, kind string) {
	if t.progressView != nil {
		switch kind {
		case "info":
			t.progressView.AddInfo(msg)
		case "success":
			t.progressView.AddSuccess(msg)
		case "warning":
			t.progressView.AddWarning(msg)
		case "error":
			t.progressView.AddError(msg)
		}
		return
	}
	switch kind {
	case "info":
		tui.PrintInfo(msg)
	case "success":
		tui.PrintSuccess(msg)
	case "warning":
		tui.PrintWarning(msg)
	case "error":
		tui.PrintError(msg)
	}
}

func (t *TUIOutput) Status(msg string) {
	if t.progressView != nil {
		t.progressView.AddStatus(msg)
		return
	}
	tui.PrintStatus(msg)
}

func (t *TUIOutput) Step(current, total int, msg string) {
	if t.progressView != nil {
		t.progressView.AddInfo(fmt.Sprintf("[%d/%d] %s", current, total, msg))
		return
	}
	tui.PrintStep(current, total, msg)
}

func (t *TUIOutput) Box(title string, lines []string) {
	if t.progressView != nil {
		if title != "" {
			t.progressView.AddText(title)
		}
		for _, line := range lines {
			t.progressView.AddText("  " + line)
		}
		return
	}
	tui.PrintBox(title, lines)
}

func (t *TUIOutput) KV(key, value string) string {
	return tui.KV(key+": ", value)
}

func (t *TUIOutput) ShowInfo(cfg actions.InfoConfig) error {
	if !t.interactive {
		// CLI mode: print as plain text
		if cfg.Title != "" {
			fmt.Println(cfg.Title)
			fmt.Println(strings.Repeat("-", len(cfg.Title)))
		}
		for _, section := range cfg.Sections {
			if section.Title != "" {
				fmt.Printf("\n%s:\n", section.Title)
			}
			for _, row := range section.Rows {
				fmt.Printf("  %-20s %s\n", row.Key+":", row.Value)
			}
		}
		return nil
	}

	tuiCfg := tui.InfoConfig{
		Title:       cfg.Title,
		Description: cfg.Description,
		CopyText:    cfg.CopyText,
	}
	for _, section := range cfg.Sections {
		tuiSection := tui.InfoSection{Title: section.Title}
		for _, row := range section.Rows {
			tuiSection.Rows = append(tuiSection.Rows, tui.InfoRow{
				Key:   row.Key,
				Value: row.Value,
			})
		}
		tuiCfg.Sections = append(tuiCfg.Sections, tuiSection)
	}
	return tui.ShowInfo(tuiCfg)
}

func (t *TUIOutput) BeginProgress(title string) {
	if !t.interactive {
		return // CLI mode: no progress view, just print normally
	}
	t.progressView = tui.NewProgressView(title)
}

func (t *TUIOutput) EndProgress() {
	if t.progressView != nil {
		t.progressView.Done()
		t.progressView = nil
	}
}

func (t *TUIOutput) DismissProgress() {
	if t.progressView != nil {
		t.progressView.Dismiss()
		t.progressView = nil
	}
}

func (t *TUIOutput) IsProgressActive() bool {
	return t.progressView != nil
}

// Separator outputs a horizontal separator line.
func (t *TUIOutput) Separator(length int) {
	if t.progressView != nil {
		t.progressView.AddText(strings.Repeat("-", length))
		return
	}
	fmt.Println(strings.Repeat("-", length))
}

var _ actions.OutputWriter = (*TUIOutput)(nil)
