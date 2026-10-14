package route

const (
	Home               = "/"
	Login              = "/login"
	Register           = "/register"
	Logout             = "/logout"
	Dashboard          = "/dashboard"
	Settings           = "/settings"
	SettingsProfile    = Settings + "/profile"
	SettingsPassword   = Settings + "/password"
	SettingsTwoFactor  = Settings + "/two-factor"
	SettingsAppearance = Settings + "/appearance"
)
