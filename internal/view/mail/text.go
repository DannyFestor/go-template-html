package mail

import (
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/a-h/templ"
)

func Text(parsed *template.Template, model any) templ.ComponentFunc {
	return func(_ context.Context, w io.Writer) error {
		err := parsed.Execute(w, model)
		if err != nil {
			return fmt.Errorf("execute the text template %s: %w", parsed.Name(), err)
		}

		return nil
	}
}
