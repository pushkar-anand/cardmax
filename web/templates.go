package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

//go:embed templates/*.html templates/partials/*.html
var templatesFS embed.FS

type (
	Renderer struct {
		templates *template.Template
	}

	Template string
	Partial  string
)

const (
	TemplateHome         = "index.html"
	TemplateCards        = "cards.html"
	TemplateRecommend    = "recommend.html"
	TemplateTransactions = "transactions.html"
)

const (
	PartialRecommendationResult Partial = "recommendation_result"
)

func GetTemplates() (*Renderer, error) {
	templates, err := template.ParseFS(templatesFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("error parsing templates: %w", err)
	}

	return &Renderer{
		templates: templates,
	}, nil
}

func (tr *Renderer) Render(w http.ResponseWriter, name Template, data map[string]any) error {
	if data == nil {
		data = make(map[string]any)
	}

	data["Year"] = time.Now().Year()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := tr.templates.ExecuteTemplate(w, string(name), data)
	if err != nil {
		return fmt.Errorf("web.Render: error executing template: %w", err)
	}

	return nil
}

func (tr *Renderer) HTMLDataHandler(name Template, data map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tr.Render(w, name, data)
		if err != nil {

		}
	}
}

func (tr *Renderer) HTMLHandler(name Template) http.HandlerFunc {
	return tr.HTMLDataHandler(name, nil)
}

// RenderPartial renders a partial template (from partials directory)
func (tr *Renderer) RenderPartial(w http.ResponseWriter, name Partial, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["Year"] = time.Now().Year()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := tr.templates.ExecuteTemplate(w, string(name), data)
	if err != nil {
		return fmt.Errorf("web.RenderPartial: error executing template: %w", err)
	}

	return nil
}
