package render

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
	"github.com/rudsonalves/quicknotes/views"
)

type RenderTemplate struct {
	session *scs.SessionManager
}

func NewRender(session *scs.SessionManager) *RenderTemplate {
	return &RenderTemplate{session: session}
}

func getTemplatePageFiles(t *template.Template, page string) (*template.Template, error) {
	return t.ParseFS(views.Files, "templates/base.html", "templates/pages/"+page)
}

func (rt *RenderTemplate) RenderPage(w http.ResponseWriter, r *http.Request, status int, page string, data any) error {
	// files := []string{
	// 	"views/templates/base.html",
	// 	"views/templates/pages/" + page}

	t := template.New("").Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
		"csrfToken": func() string {
			return csrf.Token(r)
		},
		"isAuthenticated": func() bool {
			return rt.session.Exists(r.Context(), "userId")
		},
		"userEmail": func() string {
			return rt.session.GetString(r.Context(), "userEmail")
		},
	})
	t, err := getTemplatePageFiles(t, page)
	if err != nil {
		return err
	}

	buff := &bytes.Buffer{}
	if err = t.ExecuteTemplate(buff, "base", data); err != nil {
		return err
	}
	w.WriteHeader(status)
	buff.WriteTo(w)
	return nil
}

func (rt *RenderTemplate) RenderMailBody(mailTempl string, data any) ([]byte, error) {
	t, err := template.ParseFiles("views/templates/mails/" + mailTempl)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	buff := &bytes.Buffer{}
	if err = t.Execute(buff, data); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
