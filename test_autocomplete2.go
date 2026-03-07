package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
)

func main() {
	t := textinput.New()
	t.ShowSuggestions = true
	fmt.Printf("%T\n", t.CurrentSuggestion())
}
