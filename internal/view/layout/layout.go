package layout

type Kind int

const (
	KindPublic Kind = iota
	KindApp
)

const (
	PublicMainID = "public-main"
	AppMainID    = "app-main"
)

// boostConfig swaps only the layout's own <main> on boosted navigation, so the rest keeps its DOM (ADR-0004).
func boostConfig(mainID string) string {
	return "swap:outerSync select:#" + mainID + " target:#" + mainID
}
