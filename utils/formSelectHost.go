package utils

import (
	formmmodel "sshfd/bubbles/formmodel"
	"sshfd/globals"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func generateFormSelectHost(description string, list []string) *huh.Form {
	options := []huh.Option[string]{
		huh.NewOption("(none)", ""),
	}

	for _, host := range list {
		options = append(options, huh.NewOption(host, host))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a host:").
				Description(description).
				Key("host").
				Options(
					options...,
				).Height(5),
		),
	).
		WithTheme(globals.Theme).
		WithWidth(globals.FormWidth)

	return form
}

func runFormSelectHost(description string, list []string) *huh.Form {

	form := generateFormSelectHost(description, list)
	fModel := formmmodel.NewModel(formmmodel.ModelConfig{
		Form: form,
	}).Width(globals.Width)

	tea.NewProgram(&fModel).Run()

	return form
}
