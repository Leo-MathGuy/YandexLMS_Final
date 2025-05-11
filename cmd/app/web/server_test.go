package web

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/agent"
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
	defer os.Remove("sqlite3.db")
	stop := storage.ConnectDB()
	defer close(stop)

	ctx, cancel := context.WithCancel(context.Background())
	conn := agent.StartThreads(ctx)
	defer conn.Close()
	defer cancel()

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
			t.Fatalf("unexpected error")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode == http.StatusOK {
			t.Fatalf("expected non 200 status")
		}
	})

	var token string
	if !t.Run("login1", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/login", strings.NewReader("{\"login\":\"Bob\",\"password\":\"123\"}"))
		if err != nil {
			t.Fatalf("not expected error")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("unexpected error")
		}
		if cookies := w.Result().Cookies(); len(cookies) == 0 {
			t.Fatal("no cookies recieved")
		} else {
			token = cookies[0].Value
		}
	}) {
		return
	}

	t.Run("the rest", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", "/api/v1/calculate", strings.NewReader("{\"expression\":\"2+2-(42*3)-2/4+58\",\"token\":\""+token+"\"}"))
		if err != nil {
			t.Fatalf("not expected error")
		}
		mux.ServeHTTP(w, r)

		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("unexpected error")
		}

		w = httptest.NewRecorder()
		r, err = http.NewRequest("GET", "/api/v1/expressions", nil)
		r.Header.Add("Authentication", token)
		if err != nil {
			t.Fatalf("not expected error")
		}

		mux.ServeHTTP(w, r)
		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("unexpected error")
		}

		defer w.Result().Body.Close()
		out, err := io.ReadAll(w.Result().Body)
		if err != nil {
			t.Fatalf("not expected error")
		}

		exp := "{\"expressions\":[{\"id\":1,\"result\":0,\"status\":false}]}"
		if strings.Compare(string(out), exp) != 0 {
			t.Logf("Wanted: \"%s\"", exp)
			t.Fatalf("Got:    \"%s\"", string(out))
		}

		w = httptest.NewRecorder()
		r, err = http.NewRequest("GET", "/api/v1/expressions/1", nil)
		r.Header.Add("Authentication", token)
		if err != nil {
			t.Fatalf("not expected error")
		}

		mux.ServeHTTP(w, r)
		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("unexpected error: %d", w.Result().StatusCode)
		}

		defer w.Result().Body.Close()
		out, err = io.ReadAll(w.Result().Body)
		if err != nil {
			t.Fatalf("not expected error")
		}

		exp = "{\"expression\":{\"id\":1,\"result\":0,\"status\":false}}"
		if strings.Compare(string(out), exp) != 0 {
			t.Logf("Wanted: \"%s\"", exp)
			t.Fatalf("Got:    \"%s\"", string(out))
		}
	})
}
