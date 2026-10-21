package route

const (
	Home               = "/"
	Healthz            = "/healthz"
	Readyz             = "/readyz"
	Assets             = "/assets/"
	Login              = "/login"
	ForgotPassword     = "/forgot-password"
	ResetPassword      = "/reset-password"
	Register           = "/register"
	Logout             = "/logout"
	ConfirmPassword    = "/confirm-password"
	VerifyEmail        = "/verify-email"
	VerifyEmailLink    = VerifyEmail + "/{" + VerifyEmailUserID + "}/{" + VerifyEmailHash + "}"
	VerifyEmailUserID  = "id"
	VerifyEmailHash    = "hash"
	Dashboard          = "/dashboard"
	Settings           = "/settings"
	SettingsProfile    = Settings + "/profile"
	SettingsPassword   = Settings + "/password"
	SettingsTwoFactor  = Settings + "/two-factor"
	SettingsAppearance = Settings + "/appearance"
)

const ResetPasswordTokenParam = "token"
