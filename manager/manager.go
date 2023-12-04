package manager

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
)

type Component string

// ComponentManager is a helper to manage different ViewComponents
// and being able to manage key handlers for each component
type ComponentManager struct {
	componentStack []Component
	keyHandlers    map[Component]map[tcell.Key]func()
	mutex          sync.Mutex
}

// NewComponentManager creates a new EventKeyHandler
func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		componentStack: make([]Component, 0),
		keyHandlers:    make(map[Component]map[tcell.Key]func()),
		mutex:          sync.Mutex{},
	}
}

// PushComponent adds a new component to the component stack
func (eh *ComponentManager) PushComponent(component Component) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	eh.componentStack = append(eh.componentStack, component)
}

// PopComponent removes the latest component from the component stack
func (eh *ComponentManager) PopComponent() {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.componentStack) > 0 {
		eh.componentStack = eh.componentStack[:len(eh.componentStack)-1]
	}
}

// CurrentComponent returns the current component
func (eh *ComponentManager) CurrentComponent() Component {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if len(eh.componentStack) == 0 {
		return ""
	}
  log.Debug().Msgf("Current component: %s", eh.componentStack[len(eh.componentStack)-1])
	return eh.componentStack[len(eh.componentStack)-1]
}

// SetKeyHandler sets a key handler for a specific component
func (eh *ComponentManager) SetKeyHandler(component Component, key tcell.Key, handler func()) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	if _, exists := eh.keyHandlers[component]; !exists {
		eh.keyHandlers[component] = make(map[tcell.Key]func())
	}
	eh.keyHandlers[component][key] = handler
}

func (eh *ComponentManager) SetKeyHandlerForComponent(component Component) func(key tcell.Key, handler func()) {
	return func(key tcell.Key, handler func()) {
		eh.SetKeyHandler(component, key, handler)
	}
}

// HandleKey handles a key event based on the current component
func (eh *ComponentManager) HandleKey(key tcell.Key) {
	component := eh.CurrentComponent()
	if handlers, exists := eh.keyHandlers[component]; exists {
		if handler, ok := handlers[key]; ok {
			handler()
		}
	}
}
