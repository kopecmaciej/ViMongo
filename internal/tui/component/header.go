package component

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/config"
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

	// Header is a view that displays information about the database
	// in the header of the application
	Header struct {
		*core.BaseElement
		*tview.Table

		style    *config.HeaderStyle
		baseInfo BaseInfo
	}
)

// NewHeader creates a new header view
func NewHeader() *Header {
	h := Header{
		BaseElement: core.NewBaseElement(),
		Table:       tview.NewTable(),
		baseInfo:    make(BaseInfo),
	}

	h.SetIdentifier(HeaderView)
	h.SetAfterInitFunc(h.init)

	return &h
}

func (h *Header) init() error {
	h.setStyle()

	return nil
}

func (h *Header) setStyle() {
	h.style = &h.App.GetStyles().Header
	h.Table.SetSelectable(false, false)
	h.Table.SetBorder(true)
	h.Table.SetBorderPadding(0, 0, 1, 1)
	h.Table.SetTitle(" Database Info ")
}

// SetBaseInfo sets the base information about the database
// such as status, host, port, database, version, uptime, connections, memory etc.
func (h *Header) SetBaseInfo(ctx context.Context) error {
	ss, err := h.Dao.GetServerStatus(ctx)
	if err != nil {
		return err
	}

	port := strconv.Itoa(h.Dao.Config.Port)

	orElseNil := func(i int32) string {
		if i == 0 {
			return ""
		}
		return strconv.Itoa(int(i))
	}

	h.baseInfo = BaseInfo{
		0:  {"Status", h.style.ActiveSymbol.String()},
		1:  {"Host", h.Dao.Config.Host},
		2:  {"Port", port},
		3:  {"Database", h.Dao.Config.Database},
		4:  {"Version", ss.Version},
		5:  {"Uptime", orElseNil(ss.Uptime)},
		6:  {"Connections", orElseNil(ss.CurrentConns)},
		7:  {"Available Connections", orElseNil(ss.AvailableConns)},
		8:  {"Resident Memory", orElseNil(ss.Mem.Resident)},
		9:  {"Virtual Memory", orElseNil(ss.Mem.Virtual)},
		10: {"Is Master", strconv.FormatBool(ss.Repl.IsMaster)},
	}

	return nil
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
			err := h.SetBaseInfo(ctx)
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
					return
				}
				log.Error().Err(err).Msg("Error while refreshing header")
				h.setInactiveBaseInfo(err)
				sleep += 5 * time.Second
			}
		}()
		h.App.QueueUpdateDraw(func() {
			h.Render()
		})
	}
}

// Render renders the header view
func (h *Header) Render() {
	h.Table.Clear()
	b := h.baseInfo

	maxInRow := 2
	currCol := 0
	currRow := 0

	for i := 0; i < len(b); i++ {
		if i%maxInRow == 0 && i != 0 {
			currCol += 2
			currRow = 0
		}
		order := order(i)
		h.Table.SetCell(currRow, currCol, h.keyCell(b[order].label))
		h.Table.SetCell(currRow, currCol+1, h.valueCell(b[order].value))
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

func (h *Header) keyCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text + ":")
	cell.SetTextColor(h.style.KeyColor.Color())

	return cell
}

func (h *Header) valueCell(text string) *tview.TableCell {
	cell := tview.NewTableCell(text)
	cell.SetTextColor(h.style.ValueColor.Color())

	return cell
}
