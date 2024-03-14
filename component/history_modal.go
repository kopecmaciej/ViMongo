package component

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/primitives"
)

const (
	maxHistory      = 10
	historyFilePath = "history.txt"
)

// HistoryModal is a modal with history of queries
type HistoryModal struct {
	*Component
	*primitives.ListModal

	style *config.HistoryStyle
}

func NewHistoryModal() *HistoryModal {
	h := &HistoryModal{
		Component: NewComponent("HistoryModal"),
		ListModal: primitives.NewListModal(),
	}

	h.SetAfterInitFunc(h.init)

	return h
}

// Init initializes HistoryModal
func (h *HistoryModal) init() error {
	h.setStyle()
	h.setKeybindings()

	return nil
}

func (h *HistoryModal) setStyle() {
	h.style = &h.app.Styles.History

	h.SetTitle(" History ")
	h.SetBorder(true)
	h.ShowSecondaryText(false)
	h.SetBackgroundColor(h.style.BackgroundColor.Color())
	mainStyle := tcell.StyleDefault.
		Foreground(h.style.TextColor.Color()).
		Background(h.style.BackgroundColor.Color())
	h.SetMainTextStyle(mainStyle)

	selectedStyle := tcell.StyleDefault.
		Foreground(h.style.SelectedTextColor.Color()).
		Background(h.style.SelectedBackgroundColor.Color())
	h.SetSelectedStyle(selectedStyle)
}

func (h *HistoryModal) setKeybindings() {
	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc, tcell.KeyEnter:
			eventKey := manager.EventMsg{EventKey: event, Sender: h.GetIdentifier()}
			h.app.Root.RemovePage(h.GetIdentifier())
			h.SendToComponent(InputBarComponent, eventKey)
			return nil
		}
		return event
	})
}

// Render loads history from file and renders it
func (h *HistoryModal) Render() error {
	history, err := h.loadHistory()
	if err != nil {
		return err
	}

	h.Clear()

	// load in reverse order
	for i := len(history) - 1; i >= 0; i-- {
		rune := 57 - i
		entry := history[i]
		h.AddItem(entry, "", int32(rune), nil)
	}

	h.app.Root.AddPage(h.GetIdentifier(), h, true, true)

	return nil
}

// SaveToHistory saves text to history file, if it's not already there.
// It will overwrite oldest entry if history is full.
func (h *HistoryModal) SaveToHistory(text string) error {
	history, err := h.loadHistory()

	var updatedHistory []string
	for _, line := range history {
		if line != text {
			updatedHistory = append(updatedHistory, line)
			if len(updatedHistory) >= maxHistory {
				updatedHistory = updatedHistory[1:]
			}
		}
	}
	updatedHistory = append(updatedHistory, text)

	historyFile, err := os.OpenFile(historyFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer historyFile.Close()

	for _, entry := range updatedHistory {
		_, err = historyFile.WriteString(entry + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// GetText returns text from selected item
func (h *HistoryModal) GetText() string {
	text := h.ListModal.GetText()

	return strings.TrimSpace(text)
}

// loadHistory loads history from history file
func (h *HistoryModal) loadHistory() ([]string, error) {
	file, err := os.ReadFile(historyFilePath)
	if err != nil {
		return nil, err
	}

	history := []string{}
	lines := strings.Split(string(file), "\n")

	for _, line := range lines {
		if line != "" {
			history = append(history, line)
		}
	}

	return history, nil
}
