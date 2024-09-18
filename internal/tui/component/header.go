package component

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/rs/zerolog/log"
)

const (
	HeaderView = "Header"
)

type (
	order int

	info struct {
		label string
		value string
	}

	BaseInfo map[order]info

	// Header is a view that displays basic information and keybindings in the header
	Header struct {
		*core.BaseElement
		*core.Table

		style        *config.HeaderStyle
		baseInfo     BaseInfo
		keys         []config.Key
		currentFocus tview.Identifier
	}
)

// NewHeader creates a new header view
func NewHeader() *Header {
	h := Header{
		BaseElement: core.NewBaseElement(),
		Table:       core.NewTable(),
		baseInfo:    make(BaseInfo),
	}

	h.SetIdentifier(HeaderView)
	h.SetAfterInitFunc(h.init)

	return &h
}

func (h *Header) init() error {
	h.setStyle()
	h.setStaticLayout()

	h.handleEvents()

	return nil
}

func (h *Header) setStaticLayout() {
	h.Table.SetBorder(true)
	h.Table.SetTitle(" Basic Info ")
	h.Table.SetBorderPadding(0, 0, 1, 1)
}

func (h *Header) setStyle() {
	h.style = &h.App.GetStyles().Header
	h.SetStyle(h.App.GetStyles())
}

// SetBaseInfo sets the basic information about the database connection
func (h *Header) SetBaseInfo() BaseInfo {
	h.baseInfo = BaseInfo{
		0: {"Status", h.style.ActiveSymbol.String()},
		1: {"Host", h.Dao.Config.Host},
	}
	return h.baseInfo
}

// refresh refreshes the header view every 10 seconds
// to display the most recent information about the database
func (h *Header) Refresh() {
	sleep := 10 * time.Second
	for {
		time.Sleep(sleep)
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := h.Dao.Ping(ctx)
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
					return
				}
				log.Error().Err(err).Msg("Error while refreshing header")
				h.setInactiveBaseInfo(err)
				sleep += 5 * time.Second
			}
		}()
		go h.App.QueueUpdateDraw(func() {
			h.Render()
		})
	}
}

// Render renders the header view
func (h *Header) Render() {
	h.Table.Clear()
	base := h.SetBaseInfo()

	maxInRow := 2
	currCol := 0
	currRow := 0

	for i := 0; i < len(base); i++ {
		if i%maxInRow == 0 && i != 0 {
			currCol += 2
			currRow = 0
		}
		order := order(i)
		h.Table.SetCell(currRow, currCol, h.keyCell(base[order].label))
		h.Table.SetCell(currRow, currCol+1, h.valueCell(base[order].value))
		currRow++
	}

	h.Table.SetCell(0, 2, tview.NewTableCell(" "))
	h.Table.SetCell(1, 2, tview.NewTableCell(" "))
	currCol++

	k, err := h.UpdateKeys()
	if err != nil {
		log.Warn().Err(err).Msg("Error while updating keys")
		return
	}

	for _, key := range k {
		if currRow%maxInRow == 0 && currRow != 0 {
			currCol += 2
			currRow = 0
		}
		var keyString string
		var iter []string
		// keys can be both runes and keys
		if len(key.Keys) > 0 {
			iter = append(iter, key.Keys...)
		}
		if len(key.Runes) > 0 {
			iter = append(iter, key.Runes...)
		}
		for i, k := range iter {
			if i == 0 {
				keyString = k
			} else {
				keyString = fmt.Sprintf("%s, %s", keyString, k)
			}
		}

		h.Table.SetCell(currRow, currCol, h.keyCell(keyString))
		h.Table.SetCell(currRow, currCol+1, h.valueCell(key.Description))
		currRow++
	}
}

func (h *Header) setInactiveBaseInfo(err error) {
	h.baseInfo = make(BaseInfo)
	h.baseInfo[0] = info{"Status", h.style.InactiveSymbol.String()}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
			h.baseInfo[1] = info{"Error", "Unauthorized, please check your credentials or your privileges"}
		} else {
			h.baseInfo[1] = info{"Error", err.Error()}
		}
	}
}

// handle events from the manager
func (h *Header) handleEvents() {
	go h.HandleEvents(HeaderView, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.FocusChanged:
			h.currentFocus = tview.Identifier(event.Message.Data.(tview.Identifier))
			go h.App.QueueUpdateDraw(func() {
				h.Render()
			})
		case manager.StyleChanged:
			h.setStyle()
			go h.App.QueueUpdateDraw(func() {
				h.Render()
			})
		}
	})
}

func (h *Header) keyCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text + " ")
	cell.SetTextColor(h.style.KeyColor.Color())

	return cell
}

func (h *Header) valueCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text)
	cell.SetTextColor(h.style.ValueColor.Color())

	return cell
}

// UpdateKeys updates the keybindings for the current focused element
func (h *Header) UpdateKeys() ([]config.Key, error) {
	if h.currentFocus == "" {
		return nil, nil
	}

	orderedKeys, err := h.App.GetKeys().GetKeysForElement(string(h.currentFocus))
	if err != nil {
		return nil, err
	}
	keys := orderedKeys[0].Keys

	if len(keys) > 0 {
		h.keys = keys
	} else {
		h.keys = nil
	}

	return keys, nil
}
