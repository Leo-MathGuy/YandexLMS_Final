package web

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/util"
)

func TestIndex(t *testing.T) {
	util.Leave()

	oldLogger := logging.Logger
	defer func() { logging.Logger = oldLogger }()

	logging.Logger = log.Default()

	mux := createServer()
	initServer()

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("cannot create request")
	}
	mux.ServeHTTP(w, r)

	res := w.Result()
	data, err := io.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		t.Fatalf("error %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("returned code: %s", res.Status)
	}
	if len(string(data)) < 50 {
		t.Fatalf("no page: \n%s", string(data))
	}
}

func TestServer(t *testing.T) {
	util.Leave()
	os.Remove("sqlite3.db")
	storage.ConnectDB()

	if err := storage.CreateTables(storage.D); err != nil {
		t.Fatalf("Creating tables failed with %s", err.Error())
	}

	oldLogger := logging.Logger
	defer func() { logging.Logger = oldLogger }()
	logging.Logger = log.Default()

	mux := createServer()
	initServer()

	t.Run("register no login", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"\",\"password\":\"123\"}"))
		if err != nil {
			t.Fatalf("cannot create request")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("expected error")
		}
	})

	t.Run("register long name", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"bobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbobbob\",\"password\":\"123\"}"))
		if err != nil {
			t.Fatalf("cannot create request")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("expected error")
		}
	})

	t.Run("register no password", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"bob\",\"password\":\"\"}"))
		if err != nil {
			t.Fatalf("cannot create request")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("expected error")
		}
	})

	t.Run("register long pasword", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"bob\",\"password\":\"123123123123123123123123123123123123123123123123123123123123123123123\"}"))
		if err != nil {
			t.Fatalf("cannot create request")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("expected error")
		}
	})

	if !t.Run("register", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"bob\",\"password\":\"123\"}"))
		if err != nil {
			t.Fatalf("cannot create request")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("register error: %s", w.Result().Status)
		}
	}) {
		return
	}

	t.Run("double 1", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"bob\",\"password\":\"123\"}"))
		if err != nil {
			t.Fatalf("cannot create request")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode == http.StatusOK {
			t.Fatalf("register error: %s", w.Result().Status)
		}
	})

	t.Run("double 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/register", strings.NewReader("{\"login\":\"Bob\",\"password\":\"123\"}"))
		if err != nil {
			t.Fatalf("expected error")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode == http.StatusOK {
			t.Fatalf("expected error")
		}
	})
}
