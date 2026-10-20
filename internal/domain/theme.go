package domain

import "slices"

type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

func (t Theme) Valid() bool {
	return slices.Contains([]Theme{ThemeLight, ThemeDark, ThemeSystem}, t)
}
