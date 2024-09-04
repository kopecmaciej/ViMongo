package component

import (
	"context"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/mongo"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
	"github.com/kopecmaciej/vi-mongo/internal/tui/modal"
	"github.com/kopecmaciej/vi-mongo/internal/tui/primitives"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
)

const (
	PeekerComponent = "Peeker"
)

// Peeker is a view that provides a modal view for peeking at a document
type Peeker struct {
	*core.BaseElement
	*primitives.ViewModal

	style       *config.DocPeekerStyle
	docModifier *DocModifier
	currentDoc  string

	doneFunc func()
}

// NewPeeker creates a new Peeker view
func NewPeeker() *Peeker {
	p := &Peeker{
		BaseElement: core.NewBaseElement(),
		ViewModal:   primitives.NewViewModal(),
		docModifier: NewDocModifier(),
	}

	p.SetIdentifier(PeekerComponent)
	p.SetAfterInitFunc(p.init)

	return p
}

func (p *Peeker) init() error {
	p.setStyle()
	p.setKeybindings()

	if err := p.docModifier.Init(p.App); err != nil {
		return err
	}

	return nil
}

func (p *Peeker) setStyle() {
	p.style = &p.App.GetStyles().DocPeeker
	p.SetBorder(true)
	p.SetTitle("Document Details")
	p.SetTitleAlign(tview.AlignLeft)
	p.SetHighlightColor(p.style.HighlightColor.Color())
	p.SetDocumentColors(
		p.style.KeyColor.Color(),
		p.style.ValueColor.Color(),
		p.style.BracketColor.Color(),
		p.style.ArrayColor.Color(),
	)

	p.ViewModal.AddButtons([]string{"Edit", "Close"})
}

func (p *Peeker) setKeybindings() {
	k := p.App.GetKeys()
	p.ViewModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.Peeker.MoveToTop, event.Name()):
			p.MoveToTop()
			return nil
		case k.Contains(k.Peeker.MoveToBottom, event.Name()):
			p.MoveToBottom()
			return nil
		case k.Contains(k.Peeker.CopyFullObj, event.Name()):
			if err := p.ViewModal.CopySelectedLine(clipboard.WriteAll, "full"); err != nil {
				modal.ShowError(p.App.Pages, "Error copying full line", err)
			}
			return nil
		case k.Contains(k.Peeker.CopyValue, event.Name()):
			if err := p.ViewModal.CopySelectedLine(clipboard.WriteAll, "value"); err != nil {
				modal.ShowError(p.App.Pages, "Error copying value", err)
			}
			return nil
		case k.Contains(k.Peeker.Refresh, event.Name()):
			p.setText()
			return nil
		}
		return event
	})
}

func (p *Peeker) MoveToTop() {
	p.ViewModal.MoveToTop()
}

func (p *Peeker) MoveToBottom() {
	p.ViewModal.MoveToBottom()
}

func (p *Peeker) SetDoneFunc(doneFunc func()) {
	p.doneFunc = doneFunc
}

func (p *Peeker) Render(ctx context.Context, state *mongo.CollectionState, _id interface{}) error {
	doc, err := state.GetJsonDocById(_id)
	if err != nil {
		return err
	}

	p.currentDoc = doc
	p.setText()

	p.App.Pages.AddPage(p.GetIdentifier(), p.ViewModal, true, true)
	p.ViewModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Edit" {
			updatedDoc, err := p.docModifier.Edit(ctx, state.Db, state.Coll, p.currentDoc)
			if err != nil {
				modal.ShowError(p.App.Pages, "Error editing document", err)
				return
			}

			state.UpdateRawDoc(updatedDoc)
			p.currentDoc = updatedDoc
			if p.doneFunc != nil {
				p.doneFunc()
			}
			p.setText()
			p.App.SetFocus(p.ViewModal)
		} else if buttonLabel == "Close" || buttonLabel == "" {
			p.App.Pages.RemovePage(p.GetIdentifier())
		}
	})
	return nil
}

func (p *Peeker) setText() {
	p.ViewModal.SetText(primitives.Text{
		Content: p.currentDoc,
		Color:   p.style.ValueColor.Color(),
		Align:   tview.AlignLeft,
	})
}
