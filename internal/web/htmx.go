package web

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/view/layout"
)

const (
	headerRequest     = "Hx-Request"
	headerRequestType = "Hx-Request-Type"
	headerTarget      = "Hx-Target"
	headerRetarget    = "Hx-Retarget"
	headerReselect    = "Hx-Reselect"
	headerReswap      = "Hx-Reswap"
	headerRedirect    = "Hx-Redirect"
)

func isHTMX(req *http.Request) bool {
	return req.Header.Get(headerRequest) == "true"
}

func isPartial(req *http.Request) bool {
	return req.Header.Get(headerRequestType) == "partial"
}

func crossesLayouts(req *http.Request, kind layout.Kind) bool {
	return isHTMX(req) && !targetsMain(req, layout.MainID(kind))
}

func targetsMain(req *http.Request, mainID string) bool {
	return req.Header.Get(headerTarget) == htmxElementIdentifier("main", mainID)
}

func htmxElementIdentifier(tag, id string) string {
	return tag + "#" + id
}

func swapWholeBody(header http.Header) {
	header.Set(headerRetarget, "body")
	header.Set(headerReselect, "body")
	header.Set(headerReswap, "outerSync")
}
