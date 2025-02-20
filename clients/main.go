package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	store := &Store{}
	if err := store.Init(); err != nil {
		panic(err)
	}

	m := NewModel(store)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}