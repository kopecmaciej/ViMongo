package component

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/mongui/config"
	"github.com/kopecmaciej/mongui/manager"
	"github.com/kopecmaciej/tview"
	"github.com/rs/zerolog/log"
)

const (
	ConnectorComponent = manager.Component("Connector")
)

// Connector is a view for connecting to mongodb using tview package
type Connector struct {
	*Component
	*tview.Flex

	// form is for creating new connection
	form *tview.Form

	// list is a list of all available connections
	list *tview.List

	// function that is called when connection is set
	onSubmit func()
}

// NewConnector creates a new connection view
func NewConnector() *Connector {
	c := &Connector{
		Component: NewComponent(ConnectorComponent),
		Flex:      tview.NewFlex(),
		form:      tview.NewForm(),
		list:      tview.NewList(),
	}

	return c
}

// Init overrides the Init function from the Component struct
func (c *Connector) Init(app *App) error {
	c.app = app

	c.setStyle()
	c.setKeybindings()

	c.Render()

	return nil
}

func (c *Connector) setStyle() {
	style := c.app.Styles.Connector

	c.SetBackgroundColor(style.BackgroundColor.Color())
	c.form.SetTitle(" New connection ")
	c.form.SetBorder(true)
	c.form.SetBackgroundColor(style.BackgroundColor.Color())
	c.form.SetFieldTextColor(style.FormInputColor.Color())
	c.form.SetFieldBackgroundColor(style.FormInputBackgroundColor.Color())

	c.list.SetTitle(" Saved connections ")
	c.list.SetBorder(true)
	c.list.SetBackgroundColor(style.BackgroundColor.Color())
	c.list.ShowSecondaryText(true)
	c.list.SetWrapText(true)
	c.list.SetBorderPadding(1, 1, 1, 1)
	c.list.SetItemGap(1)

	mainStyle := tcell.StyleDefault.
		Foreground(style.ListTextColor.Color()).
		Background(style.BackgroundColor.Color())
	c.list.SetMainTextStyle(mainStyle)

	secondaryStyle := tcell.StyleDefault.
		Foreground(style.ListSecondaryTextColor.Color()).
		Background(style.BackgroundColor.Color()).
		Italic(true)
	c.list.SetSecondaryTextStyle(secondaryStyle)
}

func (c *Connector) setKeybindings() {
	k := c.app.Keys
	c.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Debug().Msgf("event name %s", event.Name())
		switch {
		case k.Contains(k.Connector.ConnectorForm.SaveConnection, event.Name()):
			c.saveButtonFunc()
			return nil
		case k.Contains(k.Connector.ConnectorForm.FocusList, event.Name()):
			c.app.SetFocus(c.list)
			return nil
		}

		return event
	})

	c.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		log.Debug().Msgf("event name %s", event.Name())
		switch {
		case k.Contains(k.Connector.ConnectorList.FocusForm, event.Name()):
			c.app.SetFocus(c.form)
			return nil
		case k.Contains(k.Connector.ConnectorList.DeleteConnection, event.Name()):
			c.deleteCurrConnection()
			return nil
		}
		return event
	})
}

// Render renders the Component
func (c *Connector) Render() {
	c.Clear()

	// easy way to center the form
	c.AddItem(tview.NewBox(), 0, 1, false)

	if len(c.app.Config.Connections) > 0 {
		c.renderList()
		defer c.app.SetFocus(c.list)
	} else {
		defer c.app.SetFocus(c.form)
	}
	c.renderForm()

	// easy way to center the form
	c.AddItem(tview.NewBox(), 0, 1, false)
}

// renderForm renders the form for creating new connection
func (c *Connector) renderForm() *tview.Form {
	c.form.Clear(true)

	c.form.AddInputField("Name", "", 40, nil, nil)
	c.form.AddInputField("Url", "mongodb://", 40, nil, nil)
	c.form.AddTextView("Example", "mongodb://username:password@host:port/db", 40, 1, true, false)
	c.form.AddTextView("Info", "Provide Url or fill the form below", 41, 1, true, false)
	c.form.AddTextView(" ", "-- ----------------------------------------", 40, 1, true, false)
	c.form.AddInputField("Host", "", 40, nil, nil)
	c.form.AddInputField("Port", "", 10, nil, nil)
	c.form.AddInputField("Username", "", 40, nil, nil)
	c.form.AddPasswordField("Password", "", 40, '*', nil)
	c.form.AddInputField("Database", "", 40, nil, nil)
	c.form.AddInputField("Timeout", "5", 10, nil, nil)

	c.form.AddButton("Save", c.saveButtonFunc)
	c.form.AddButton("Cancel", c.cancelButtonFunc)

	c.AddItem(c.form, 60, 0, true)

	return c.form
}

// renderList renders the list of all available connections
func (c *Connector) renderList() {
	c.list.Clear()

	for _, conn := range c.app.Config.Connections {
		uri := "uri: " + conn.GetUri()
		c.list.AddItem(conn.Name, uri, 0, func() {
			c.setConnections()
		})
	}

	c.list.AddItem("Click to add new connection", "or by pressing "+c.app.Keys.Connector.ConnectorList.FocusForm.String(), 0, func() {
		c.app.SetFocus(c.form)
	})

	c.AddItem(c.list, 50, 0, true)
}

// setConnections sets connections from config file
func (c *Connector) setConnections() {
	// if last item is "Click to add new connection" then return
	log.Debug().Msgf("item count %d", c.list.GetItemCount())
	log.Debug().Msgf("current item %d", c.list.GetCurrentItem())
	if c.list.GetCurrentItem() == c.list.GetItemCount()-1 {
		return
	}
	connName, _ := c.list.GetItemText(c.list.GetCurrentItem())
	err := c.app.Config.SetCurrentConnection(connName)
	if err != nil {
		ShowErrorModal(c.app.Root, "Failed to set current connection", err)
		return
	}
	c.app.Config.CurrentConnection = connName
	if c.onSubmit != nil {
		c.onSubmit()
	}
}

// removeConnection removes connection from config file
func (c *Connector) deleteCurrConnection() error {
	currItem := c.list.GetCurrentItem()
	currConn, _ := c.list.GetItemText(currItem)
	err := c.app.Config.DeleteConnection(currConn)
	if err != nil {
		return err
	}

	c.Render()
	c.list.SetCurrentItem(currItem)

	return nil
}

// saveButtonFunc is a function for saving new connection
func (c *Connector) saveButtonFunc() {
	name := c.form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
	url := c.form.GetFormItemByLabel("Url").(*tview.InputField).GetText()
	timeout := c.form.GetFormItemByLabel("Timeout").(*tview.InputField).GetText()
	intTimeout, err := strconv.Atoi(timeout)
	if err != nil {
		ShowErrorModal(c.app.Root, "Timeout must be a number", err)
		return
	}
	if url != "mongodb://" {
		if name == "" {
			name = url
		}
		err := c.app.Config.AddConnectionFromUri(&config.MongoConfig{
			Name:    name,
			Uri:     url,
			Timeout: intTimeout,
		})
		if err != nil {
			ShowErrorModal(c.app.Root, "Failed to save connection", err)
			c.form.GetFormItemByLabel("Name").(*tview.InputField).SetText("")
			return
		}
	} else {
		host := c.form.GetFormItemByLabel("Host").(*tview.InputField).GetText()
		port := c.form.GetFormItemByLabel("Port").(*tview.InputField).GetText()
		intPort, err := strconv.Atoi(port)
		if err != nil {
			ShowErrorModal(c.app.Root, "Port must be a number", err)
			return
		}
		username := c.form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := c.form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		database := c.form.GetFormItemByLabel("Database").(*tview.InputField).GetText()

		if name == "" {
			name = host + ":" + port
		}
		err = c.app.Config.AddConnection(&config.MongoConfig{
			Name:     name,
			Host:     host,
			Port:     intPort,
			Username: username,
			Password: password,
			Database: database,
			Timeout:  intTimeout,
		})
		if err != nil {
			ShowErrorModal(c.app.Root, "Failed to save connection", err)
			return
		}
	}
	c.Render()
	c.list.SetCurrentItem(c.list.GetItemCount())
}

// cancelButtonFunc is a function for canceling the form
func (c *Connector) cancelButtonFunc() {
	c.renderForm()
}

// SetOnSubmitFunc sets callback function
func (c *Connector) SetOnSubmitFunc(onSubmit func()) {
	c.onSubmit = onSubmit
}
