package navigation

type actionKey int

const (
	Main actionKey = iota
	Installer
	Shortcut
	OSK
)

type Navigator struct {
	actionRegistry map[actionKey]func()
	// add shared state fields here
}

func NewNavigator() *Navigator {
	n := &Navigator{
		actionRegistry: make(map[actionKey]func()),
	}

	// registering menu handlers
	n.actionRegistry = map[actionKey]func(){}

	return n
}
