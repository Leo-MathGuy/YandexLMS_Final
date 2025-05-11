package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/util"
)

func apiTester(
	t *testing.T,
	method string,
	body any,
	_ string,
	url string,
	f func(http.ResponseWriter, *http.Request),
	expect bool,
	bodyRaw bool,
) {
	var b []byte
	if method == http.MethodPost {
		if bodyRaw {
			b = body.([]byte)
		} else {
			var err error
			b, err = json.Marshal(body)
			if err != nil {
				t.Fatalf("Error marshaling request: %s", err)
			}
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("Error making request: %s", err)
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(f).ServeHTTP(rr, req)

	response := rr.Body.Bytes()

	if status := rr.Code; (status == http.StatusOK) != expect && strconv.FormatInt(int64(status), 10)[0] != '5' {
		if b == nil {
			b = make([]byte, 0)
		}
		t.Fatalf("Wrong status: %d\nData: %s\nResponse: %s", status, string(b), string(response))
	}
}

func TestAPI(t *testing.T) {
	// May god forgive me for this abomination
	util.Leave()
	os.Remove("testapi.db")
	defer os.Remove("testapi.db")
	t.Setenv("APPDB", "./testapi.db")
	defer t.Setenv("APPDB", "")
	defer storage.DisconnectDB()
	stop := storage.ConnectDB()
	storage.CreateTables(storage.D)
	defer close(stop)

	t.Run("favicon", func(t *testing.T) {
		apiTester(t, http.MethodGet, nil, "", "/favicon.ico", Favicon, true, false)
	})

	type AuthTest struct {
		pass bool
		AuthRequest
	}

	registerTest := []AuthTest{
		{true, AuthRequest{"bob", "123"}},
		{true, AuthRequest{"Eve", "1kfkafl23"}},
		{true, AuthRequest{"Alic3", "121xd24@DR@$"}},
		{false, AuthRequest{"3ve", "341fd"}},
		{false, AuthRequest{"Candice", ""}},
		{false, AuthRequest{"Candice", "34"}},
		{false, AuthRequest{"Candice", strings.Repeat("abc", 14)}},
		{false, AuthRequest{"Alic$", "password"}},
	}

	registerTest2 := []AuthTest{
		{false, AuthRequest{"Bob", "password"}},
		{false, AuthRequest{"eve", "password"}},
	}

	if !t.Run("Register", func(t *testing.T) {
		wg := sync.WaitGroup{}

		if !t.Run("Phase 1", func(t *testing.T) {
			for _, test := range registerTest {
				wg.Add(1)
				go func() {
					defer wg.Done()
					apiTester(t, http.MethodPost, test.AuthRequest, "", "/api/v1/register", RegisterAPI, test.pass, false)
				}()
			}
			wg.Wait()

		}) {
			return
		}

		for _, test := range registerTest2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				apiTester(t, http.MethodPost, test.AuthRequest, "", "/api/v1/register", RegisterAPI, test.pass, false)
			}()
		}
		wg.Wait()

		apiTester(t, http.MethodPost, []byte("test"), "", "/api/v1/register", RegisterAPI, false, true)
	}) {
		return
	}

	loginTest := []AuthTest{
		{true, AuthRequest{"bob", "123"}},
		{true, AuthRequest{"Bob", "123"}},
		{true, AuthRequest{"boB", "123"}},
		{true, AuthRequest{"alIC3", "121xd24@DR@$"}},
		{false, AuthRequest{"bob", ""}},
		{false, AuthRequest{"bob", strings.Repeat("-", 1000)}},
		{false, AuthRequest{"bob", "12334"}},
		{false, AuthRequest{"eve", "password"}},
	}

	if !t.Run("Login", func(t *testing.T) {
		wg := sync.WaitGroup{}

		for _, test := range loginTest {
			wg.Add(1)
			go func() {
				defer wg.Done()
				apiTester(t, http.MethodPost, test.AuthRequest, "", "/api/v1/login", LoginAPI, test.pass, false)
			}()
		}
		wg.Wait()

		apiTester(t, http.MethodPost, []byte("test"), "", "/api/v1/login", LoginAPI, false, true)
	}) {
		return
	}

	type ExprTest struct {
		pass bool
		CalcRequest
	}

	token, err := storage.CreateToken("bob")
	if err != nil {
		t.Fatalf("Failed to create token: %s", err)
	}

	exprTests := []ExprTest{
		{true, CalcRequest{"2+2", token}},
		{true, CalcRequest{"2+2-(2*5-2)", token}},
		{false, CalcRequest{"2+", token}},
		{false, CalcRequest{"2+2", ""}},
	}

	if !t.Run("Expressions", func(t *testing.T) {
		for _, test := range exprTests {
			apiTester(t, http.MethodPost, test.CalcRequest, "", "/api/v1/calculate", Calculate, test.pass, false)
		}
		apiTester(t, http.MethodPost, []byte("test"), "", "/api/v1/calculate", Calculate, false, true)
	}) {
		return
	}
}

func TestWeb(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Templates error: %s", r)
		}
	}()
	CheckTemplates()
}
