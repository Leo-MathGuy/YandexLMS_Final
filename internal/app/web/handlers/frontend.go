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

// Wizardry was involved
func parseTemplate(page string) *template.Template {
	util.Leave()

	templ := template.New("")
	err := filepath.Walk("./web/components", func(path string, info os.FileInfo, err error) error {
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

	_, err = templ.ParseFiles("./web/pages/" + page)
	if err != nil {
		panic("Error parsing templates: " + err.Error())
	}

	return templ
}

func CheckTemplates() {
	util.Leave()

	err := filepath.Walk("./web/pages", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			parseTemplate(info.Name())
		}

		return err
	})
	if err != nil {
		panic("Error checking templates: " + err.Error())
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	if err := parseTemplate("index.html").ExecuteTemplate(w, "index.html", nil); err != nil {
		logging.Error("Failure to render template: %s", err.Error())
	}
}

func Sans(w http.ResponseWriter, r *http.Request) {
	if err := parseTemplate("sans.html").ExecuteTemplate(w, "sans.html", nil); err != nil {
		logging.Error("Failure to render template: %s", err.Error())
	}
}

type authData struct {
	Link     string
	Name     string
	ShowInfo bool
}

func renderCheck(w http.ResponseWriter, err error) {
	if err != nil {
		logging.Error("Failure to render template: %s", err.Error())
		http.Error(w, "Render failure", http.StatusInternalServerError)
	}
}

func renderPage(w http.ResponseWriter, page string, data any) error {
	return parseTemplate(page).ExecuteTemplate(w, page, data)
}

func checkCookie(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("token")
	if err == nil {
		if u, err := storage.CheckToken(storage.D, cookie.Value); err == nil && u != nil {
			return true
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	return false
}

func Register(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		renderCheck(w, renderPage(w, "auth.html", authData{"register", "Register", true}))
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		renderCheck(w, renderPage(w, "auth.html", authData{"login", "Log In", false}))
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func Calc(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) {
		renderCheck(w, renderPage(w, "calc.html", nil))
		return
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
