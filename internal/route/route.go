package route

const (
	Home               = "/"
	Healthz            = "/healthz"
	Readyz             = "/readyz"
	Assets             = "/assets/"
	Login              = "/login"
	ForgotPassword     = "/forgot-password"
	Register           = "/register"
	Logout             = "/logout"
	Dashboard          = "/dashboard"
	Settings           = "/settings"
	SettingsProfile    = Settings + "/profile"
	SettingsPassword   = Settings + "/password"
	SettingsTwoFactor  = Settings + "/two-factor"
	SettingsAppearance = Settings + "/appearance"
)
