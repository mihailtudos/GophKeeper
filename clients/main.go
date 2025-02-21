package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	store := &Store{}
	if err := store.Init(); err != nil {
		panic(err)
	}

	if err := storeAuthCreds("test1234"); err != nil {
		panic(err)
	}

	token, err := getAuthCreds()
	if err != nil {
		panic(err)
	}

	fmt.Println(token)

	m := NewModel(store)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
