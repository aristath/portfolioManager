package ui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	Quit         key.Binding
	Back         key.Binding
	OpenSettings key.Binding
	SaveSettings key.Binding
}

var keys = keyMap{
	Quit:         key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Back:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	OpenSettings: key.NewBinding(key.WithKeys("s", "o"), key.WithHelp("s/o", "settings")),
	SaveSettings: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
}
