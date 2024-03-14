package component

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/mongui/mongo"
	"github.com/rivo/tview"
)

type (
	// App is a main application struct
	App struct {
		*tview.Application

		Dao     *mongo.Dao
		Manager *manager.ComponentManager
		Root    *Root
		Styles  *config.Styles
		Config  *config.Config
		Keys    *config.KeyBindings
	}
)

func NewApp(appConfig *config.Config) App {
	styles := config.NewStyles()
	keyBindings := config.NewKeyBindings()

	app := App{
		Application: tview.NewApplication(),
		Root:        NewRoot(),
		Manager:     manager.NewComponentManager(),
		Styles:      styles,
		Config:      appConfig,
		Keys:        &keyBindings,
	}

	return app
}

// Init initializes app
func (a *App) Init() error {
	ctx := context.Background()
	a.Root.app = a
	if err := a.Root.Init(); err != nil {
		return err
	}
	a.SetRoot(a.Root.Pages, true).EnableMouse(true)

	help := NewHelp()
	err := help.Init(a)
	if err != nil {
		return err
	}
	a.setKeybindings(ctx, help)

	return a.Run()
}

func (a *App) setKeybindings(ctx context.Context, help *Help) {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case a.Keys.Contains(a.Keys.Global.ToggleHelp, event.Name()):
			if a.Root.HasPage(string(HelpComponent)) {
				a.Root.RemovePage(HelpComponent)
				return nil
			}
			err := help.Render()
			if err != nil {
				return event
			}
			a.Root.AddPage(HelpComponent, help, true, true)
			return nil
		}
		return event
	})
}
