package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

//go:embed templates/*.html.tmpl templates/partials/*.tmpl
var templatesFS embed.FS

type (
	Renderer struct {
		templates *template.Template
	}

	Template string
)

const (
	TemplateHome         = "index.html"
	TemplateCards        = "cards.html"
	TemplateRecommend    = "recommend.html"
	TemplateTransactions = "transactions.html"
)

func GetTemplates() (*Renderer, error) {
	templates, err := template.ParseFS(templatesFS, "templates/*.html.tmpl", "templates/partials/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing templates: %w", err)
	}

	return &Renderer{
		templates: templates,
	}, nil
}

func (tr *Renderer) Render(w http.ResponseWriter, name Template, data map[string]any) error {
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
