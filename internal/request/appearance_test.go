package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

func TestParseAppearanceReadsTheTrimmedTheme(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		request.Appearance{Theme: domain.ThemeDark},
		request.ParseAppearance(url.Values{"theme": {" dark\n"}}),
	)
}

func TestAnAppearanceWithAKnownThemeHasNoErrors(t *testing.T) {
	t.Parallel()

	for _, theme := range []string{"light", "dark", "system"} {
		assert.Empty(t, request.ParseAppearance(url.Values{"theme": {theme}}).Validate(), theme)
	}
}

func TestAnAppearanceWithAnUnknownOrMissingThemeNamesTheThemeField(t *testing.T) {
	t.Parallel()

	for _, form := range []url.Values{{"theme": {"sepia"}}, {"theme": {"Dark"}}, {}} {
		assert.Equal(t,
			validate.Errors{"theme": "Choose light, dark or system."},
			request.ParseAppearance(form).Validate(),
			form,
		)
	}
}
