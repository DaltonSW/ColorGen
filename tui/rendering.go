package tui

import (
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var OverallRender = lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)
var ViewportBorder = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4400FF"))
var TitleStyle = lipgloss.NewStyle().Underline(true).Bold(true)

func GetTermSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	return width, height
}

func CenterHorz(msg string) string {
	w, _ := GetTermSize()

	return lipgloss.PlaceHorizontal(w, lipgloss.Center, msg)

}

func Center(msg string) string {
	w, h := GetTermSize()

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, msg)
}