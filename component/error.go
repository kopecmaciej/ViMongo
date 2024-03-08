package component

import (
	"github.com/kopecmaciej/mongui/manager"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const (
	ErrorComponent manager.Component = "Error"
)

// ShowErrorModal shows a modal with an error message
// and logs the error if it's passed
func ShowErrorModal(page *Root, message string, err error) {
	if err != nil {
		log.Error().Err(err)
	}

	message = message + "\n\n" + "For more information check the logs"

	errModal := tview.NewModal()
	errModal.SetTitle(" Error ")
	errModal.SetBorderPadding(0, 0, 1, 1)
	errModal.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	errModal.SetText(message)
	errModal.AddButtons([]string{"Ok"})

	errModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Ok" {
			page.RemovePage(ErrorComponent)
		}
	})
	page.AddPage(ErrorComponent, errModal, true, true)
}
