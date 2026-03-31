package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
)

type Session struct {
	ID        string
	Project   string
	GitURL    string
	Tool      string
	CreatedAt time.Time
}

type Choice struct {
	Label string
	Value string
}

func PromptURL() (string, error) {
	var url string
	err := huh.NewInput().
		Title("Enter git URL").
		Value(&url).
		Run()
	return url, err
}

func SelectRemote(remotes []string) (string, error) {
	var selected string
	opts := make([]huh.Option[string], len(remotes))
	for i, r := range remotes {
		opts[i] = huh.NewOption(r, r)
	}
	err := huh.NewSelect[string]().
		Title("Select remote").
		Options(opts...).
		Value(&selected).
		Run()
	return selected, err
}

func SelectSession(sessions []Session) (string, error) {
	if len(sessions) == 0 {
		return "", fmt.Errorf("no sessions available")
	}

	var selectedID string
	opts := make([]huh.Option[string], len(sessions))
	for i, s := range sessions {
		var label string
		if s.Project != "" {
			label = fmt.Sprintf("%s (%s) • %s", s.Project, s.Tool, s.CreatedAt.Format("2006-01-02 15:04"))
		} else {
			label = fmt.Sprintf("%s (%s) • %s", s.GitURL, s.Tool, s.CreatedAt.Format("2006-01-02 15:04"))
		}
		opts[i] = huh.NewOption(label, s.ID)
	}

	err := huh.NewSelect[string]().
		Title("Select a session to restore").
		Options(opts...).
		Value(&selectedID).
		Run()

	return selectedID, err
}

func SelectChoice(title string, choices []Choice) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no options available")
	}

	var selected string
	opts := make([]huh.Option[string], len(choices))
	for i, choice := range choices {
		opts[i] = huh.NewOption(choice.Label, choice.Value)
	}

	err := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(&selected).
		Run()

	return selected, err
}

func PromptText(title, initialValue string) (string, error) {
	value := initialValue
	err := huh.NewInput().
		Title(title).
		Value(&value).
		Run()

	return value, err
}

func ConfirmChoice(title string, defaultValue bool) (bool, error) {
	choices := []Choice{
		{Label: "Yes", Value: "yes"},
		{Label: "No", Value: "no"},
	}

	if !defaultValue {
		choices[0], choices[1] = choices[1], choices[0]
	}

	selected, err := SelectChoice(title, choices)
	if err != nil {
		return false, err
	}

	return selected == "yes", nil
}
