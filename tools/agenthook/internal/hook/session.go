package hook

// Session handles the hook events of one agent session. Every rule judges the checkout of the path it is about,
// never the directory the hook runs in.
type Session struct {
	id      string
	environ []string
}

func NewSession(id string, environ []string) Session {
	return Session{id: id, environ: environ}
}
