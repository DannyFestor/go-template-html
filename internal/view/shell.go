package view

type Shell struct {
	AppName     string
	Script      string
	Stylesheets []string
	User        ShellUser
}

type ShellUser struct {
	Name     string
	Email    string
	Initials string
}
