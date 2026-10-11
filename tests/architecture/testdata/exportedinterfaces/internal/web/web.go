package web

import "io"

type Renderer interface {
	Render(w io.Writer) error
}

type (
	component interface{ Render(w io.Writer) error }
	Number    interface{ ~int | ~int64 }
)

type Responder struct{}
