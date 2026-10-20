package view

type Shell struct {
	AppName     string
	Script      string
	Stylesheets []string
	User        ShellUser
	Status      string
	Theme       string
}

type ShellUser struct {
	Name     string
	Email    string
	Initials string
}
