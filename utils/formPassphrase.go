package utils

import (
	formmmodel "github.com/demingongo/sshfd/bubbles/formmodel"
	"github.com/demingongo/sshfd/globals"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func generateFormPassphrase(description string) *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				EchoMode(huh.EchoModePassword).
				Title("Passphrase:").
				Description(description).
				Key("passphrase"),
		),
	).
		WithTheme(globals.Theme).
		WithWidth(globals.FormWidth)

	return form
}

func runFormPassphrase(description string) *huh.Form {

	form := generateFormPassphrase(description)
	fModel := formmmodel.NewModel(formmmodel.ModelConfig{
		Form: form,
	}).Width(globals.Width)

	tea.NewProgram(&fModel).Run()

	return form
}
