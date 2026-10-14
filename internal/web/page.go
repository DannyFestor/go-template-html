package web

import (
	"context"
	"io"

	"github.com/DannyFestor/go-template-html/internal/view/layout"
)

type component interface {
	Render(ctx context.Context, w io.Writer) error
}

type Page struct {
	Layout   layout.Kind
	Body     component
	Fragment component
}
