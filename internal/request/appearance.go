package request

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

type Appearance struct {
	Theme domain.Theme
}

func ParseAppearance(form url.Values) Appearance {
	return Appearance{Theme: domain.Theme(Field(form, FieldTheme))}
}

func (a Appearance) Validate() validate.Errors {
	errors := validate.Errors{}

	errors.Check(a.Theme.Valid(), FieldTheme, "Choose light, dark or system.")

	return errors
}
