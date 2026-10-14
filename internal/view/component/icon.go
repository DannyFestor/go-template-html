package component

type IconName string

const (
	IconNone          IconName = ""
	IconDashboard     IconName = "dashboard"
	IconSettings      IconName = "settings"
	IconMenu          IconName = "menu"
	IconChevronDown   IconName = "chevron-down"
	IconChevronUpDown IconName = "chevron-up-down"
	IconLogout        IconName = "logout"
)

func heroiconOutlinePath(name IconName) string {
	switch name {
	case IconDashboard:
		return "M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 " +
			"2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 " +
			"2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 " +
			"0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 " +
			"0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z"
	case IconSettings:
		return "M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 " +
			"0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 " +
			"0m-9.75 0h9.75"
	case IconMenu:
		return "M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
	case IconChevronDown:
		return "m19.5 8.25-7.5 7.5-7.5-7.5"
	case IconChevronUpDown:
		return "M8.25 15 12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9"
	case IconLogout:
		return "M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 " +
			"2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9"
	case IconNone:
		return ""
	}

	return ""
}
