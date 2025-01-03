package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.dalton.dog/colorgen/config"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle
)

var ControlOrder []string = []string{
	"Bold",
	"Under",
	"Blink",
	"Fore",
	"Back",
	"Files",
	"Save",
	"Discard",
}

type StyleModel struct {
	Theme string `yaml:"theme"`

	StyleName string `yaml:"name"`
	NameInput textinput.Model

	Bold  bool `yaml:"bold"`
	Under bool `yaml:"under"`
	Blink bool `yaml:"blink"`

	ForeColor int `yaml:"fore"`
	BackColor int `yaml:"back"`

	FileArea  textarea.Model
	FileTypes []string `yaml:"filetypes"`

	Focused int

	ViewWidth  int
	ViewHeight int
}

func NewStyleEditModel(themeName, styleName string, viewWidth, viewHeight int) StyleModel {
	nameInput := textinput.New()
	// nameInput.Focus()
	nameInput.Placeholder = "Style Name"
	nameInput.SetValue(styleName)
	nameInput.PromptStyle = focusedStyle
	nameInput.TextStyle = focusedStyle
	nameInput.Prompt = ""

	newStyle := config.LoadStyle(themeName, styleName)
	if newStyle != nil {
		fileArea := textarea.New()
		fileArea.SetValue(strings.Join(newStyle.FileTypes, "\n"))
		return StyleModel{
			Theme:     themeName,
			StyleName: styleName,
			Bold:      newStyle.Bold,
			Under:     newStyle.Under,
			Blink:     newStyle.Blink,
			ForeColor: newStyle.ForeColor,
			BackColor: newStyle.BackColor,
			FileTypes: newStyle.FileTypes,
			FileArea:  fileArea,
			NameInput: nameInput,
		}

	}
	foreSlider := 128
	backSlider := -1

	fileArea := textarea.New()
	fileArea.Placeholder = ".mp3\n.gif\n.docx\n..."
	fileArea.Blur()

	return StyleModel{
		Theme:      themeName,
		NameInput:  nameInput,
		ForeColor:  foreSlider,
		BackColor:  backSlider,
		ViewWidth:  viewWidth,
		ViewHeight: viewHeight,
		FileArea:   fileArea,
	}
}

func (m StyleModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink)
}

func (m StyleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var outCmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "esc", "q":
		// 	if msg.String() == "q" && m.FileArea.Focused() {
		// 		break
		// 	}
		// 	return NewThemeModel(m.Theme, m.ViewWidth, m.ViewHeight), nil
		case "tab", "down", "shift+tab", "up":
			if msg.String() == "tab" || msg.String() == "down" {

				m.Focused++
				if m.Focused >= len(ControlOrder) {
					m.Focused = 0
				}
			} else {
				m.Focused--
				if m.Focused < 0 {
					m.Focused = len(ControlOrder) - 1
				}
			}

			if ControlOrder[m.Focused] == "Files" {
				outCmds = append(outCmds, m.FileArea.Focus())
			} else {
				m.FileArea.Blur()
			}
		case "space", "enter":
			// Toggle checkboxes if focused
			switch ControlOrder[m.Focused] {
			case "Bold":
				m.Bold = !m.Bold
			case "Under":
				m.Under = !m.Under
			case "Blink":
				m.Blink = !m.Blink
			case "Save":
				var styleToSave = config.Style{
					Theme:     m.Theme,
					StyleName: m.NameInput.Value(),
					Bold:      m.Bold,
					Under:     m.Under,
					Blink:     m.Blink,
					ForeColor: m.ForeColor,
					BackColor: m.BackColor,
					FileTypes: strings.Split(m.FileArea.Value(), "\n"),
				}
				config.SaveStyle(styleToSave)
				return NewThemeModel(m.Theme, m.ViewWidth, m.ViewHeight), nil
			case "Discard":
				return NewThemeModel(m.Theme, m.ViewWidth, m.ViewHeight), nil
			}
		case "left", "right":
			// Adjust sliders
			if ControlOrder[m.Focused] == "Fore" {
				m.ForeColor = clamp(m.ForeColor+sliderAdjustment(msg.String()), -1, 255)
			} else if ControlOrder[m.Focused] == "Back" {
				m.BackColor = clamp(m.BackColor+sliderAdjustment(msg.String()), -1, 255)
			}
		}
	case tea.WindowSizeMsg:
		m.ViewWidth = msg.Width
		m.ViewHeight = msg.Height

	}

	var cmd tea.Cmd
	m.NameInput, cmd = m.NameInput.Update(msg)
	outCmds = append(outCmds, cmd)

	m.FileArea, cmd = m.FileArea.Update(msg)
	outCmds = append(outCmds, cmd)

	return m, tea.Batch(outCmds...)
}

func (m StyleModel) View() string {
	return Center(fmt.Sprintf(
		"%v\n\n"+
			"%v\n\n"+
			"%v\n\n"+
			"%v\n\n"+
			"%v\n\n"+
			"%v\n\n",
		m.renderName(),
		m.renderStyles(),
		m.renderSliders(),
		m.renderPreview(),
		m.renderFileArea(),
		m.renderButtons()))
}

func (m StyleModel) renderName() string {
	return CenterHorz(fmt.Sprintf("%v\n\n%v", TitleStyle.Render("Name"), m.NameInput.View()))
}

func (m StyleModel) renderStyles() string {
	sOut := TitleStyle.Render("Styles") + "\n\n" +
		"%s\n" +
		"%s\n" +
		"%s\n\n"

	return CenterHorz(fmt.Sprintf(sOut,
		checkboxView(m.Bold, "Bold     ", ControlOrder[m.Focused] == "Bold"),
		checkboxView(m.Under, "Underline", ControlOrder[m.Focused] == "Under"),
		checkboxView(m.Blink, "Blink    ", ControlOrder[m.Focused] == "Blink")))
}

func (m StyleModel) renderSliders() string {
	var foreStr string
	var backStr string
	if ControlOrder[m.Focused] == "Fore" {
		foreStr = focusedStyle.Render(" > Foreground:") + " %v [%03d]"
		backStr = "   Background: %v [%03d]"
	} else if ControlOrder[m.Focused] == "Back" {
		foreStr = "   Foreground: %v [%03d]"
		backStr = focusedStyle.Render(" > Background:") + " %v [%03d]"
	} else {
		foreStr = "   Foreground: %v [%03d]"
		backStr = "   Background: %v [%03d]"
	}
	foreStr = fmt.Sprintf(foreStr, renderSlider(m.ForeColor, m.ViewWidth/2), m.ForeColor)
	backStr = fmt.Sprintf(backStr, renderSlider(m.BackColor, m.ViewWidth/2), m.BackColor)

	return CenterHorz(fmt.Sprintf(TitleStyle.Render("Colors")+"\n\n%v\n%v", foreStr, backStr))
}

func (m StyleModel) renderPreview() string {
	var backColor lipgloss.Color
	if m.BackColor == -1 {
		backColor = lipgloss.Color("")
	} else {
		backColor = lipgloss.Color(strconv.Itoa(m.BackColor))
	}
	previewColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color(strconv.Itoa(m.ForeColor))).
		Background(backColor).
		Bold(m.Bold).Underline(m.Under).Blink(m.Blink)

	return CenterHorz(TitleStyle.Render("Preview") + "\n\n" + previewColor.Render("file.example"))
}

func (m StyleModel) renderFileArea() string {
	return CenterHorz(fmt.Sprintf("%v\n%v", TitleStyle.Render("File Types"), m.FileArea.View()))
}

func (m StyleModel) renderButtons() string {
	var save string
	var discard string
	if ControlOrder[m.Focused] == "Save" {
		save = focusedStyle.Render("[ Save & Exit ]")
		discard = "[ " + blurredStyle.Render("Discard & Exit") + " ]"
	} else if ControlOrder[m.Focused] == "Discard" {
		save = "[ " + blurredStyle.Render("Save & Exit") + " ]"
		discard = focusedStyle.Render("[ Discard & Exit ]")
	} else {
		save = "[ " + blurredStyle.Render("Save & Exit") + " ]"
		discard = "[ " + blurredStyle.Render("Discard & Exit") + " ]"

	}
	return CenterHorz(fmt.Sprintf("%v\n\n%v\n\n", save, discard))
}

func (m StyleModel) GetDirColorBlock() string {
	outStr := " # " + m.StyleName + "\n\n"

	styleStr := ""

	if m.Bold {
		styleStr += "1;"
	}

	if m.Under {
		styleStr += "4;"
	}

	if m.Blink {
		styleStr += "5;"
	}

	if m.ForeColor != -1 {
		styleStr += fmt.Sprintf("38;5;%v;", strconv.Itoa(m.ForeColor))
	}

	if m.BackColor != -1 {
		styleStr += fmt.Sprintf("48;5;%v;", strconv.Itoa(m.BackColor))
	}

	styleStr = strings.TrimSuffix(styleStr, ";")

	for _, file := range m.FileTypes {
		outStr += fmt.Sprintf("%v %v\n", file, styleStr)
	}

	return outStr
}

func renderSlider(value, width int) string {
	bar := ""
	totalBlocks := width
	position := value * totalBlocks / 255
	for i := 0; i < totalBlocks; i++ {
		if i <= position {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func checkboxView(checked bool, label string, focused bool) string {
	var check string
	var focus string
	var style lipgloss.Style
	if checked {
		check = "[x]"
	} else {
		check = "[ ]"
	}

	if focused {
		focus = " >"
		style = focusedStyle
	} else {
		focus = "  "
		style = noStyle
	}

	return style.Render(fmt.Sprintf("%v %v %v", focus, check, label))
}

func sliderAdjustment(key string) int {
	if key == "right" {
		return 1
	} else if key == "left" {
		return -1
	}
	return 0
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
