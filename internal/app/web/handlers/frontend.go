package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/util"
)

func ParseTemplates() *template.Template {
	util.Leave()

	templ := template.New("")
	err := filepath.Walk("./web", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				panic("Error parsing templates: " + err.Error())
			}
		}

		return err
	})

	if err != nil {
		panic(err)
	}

	return templ
}

var t = ParseTemplates()

func Index(w http.ResponseWriter, r *http.Request) {
	if err := t.ExecuteTemplate(w, "index.html", nil); err != nil {
		logging.Error("Failure to render template: %s", err.Error())
	}
}

type authData struct {
	Link     string
	Name     string
	ShowInfo bool
}

func Register(w http.ResponseWriter, r *http.Request) {
	if err := t.ExecuteTemplate(w, "auth.html", authData{"register", "Register", true}); err != nil {
		logging.Error("Failure to render template: %s", err.Error())
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if err := t.ExecuteTemplate(w, "auth.html", authData{"login", "Log In", false}); err != nil {
		logging.Error("Failure to render template: %s", err.Error())
	}
}

func Calc(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err == nil {
		if u, err := storage.CheckToken(storage.D, cookie.Value); err == nil && u != nil {
			t.ExecuteTemplate(w, "calc.html", nil)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}
