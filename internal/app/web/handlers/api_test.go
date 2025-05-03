package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	url string,
	f func(http.ResponseWriter, *http.Request),
	expect bool,
) {
	var b []byte
	if method == http.MethodPost {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Error marshaling request: %s", err)
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
	util.Leave()
	t.Setenv("APPDB", ":memory:")
	defer t.Setenv("APPDB", "")
	defer storage.DisconnectDB()
	stop := storage.ConnectDB()
	storage.CreateTables(storage.D)
	defer close(stop)

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

	t.Run("Register", func(t *testing.T) {
		wg := sync.WaitGroup{}

		if !t.Run("Phase 1", func(t *testing.T) {

			for _, passingTest := range registerTest {
				wg.Add(1)
				go func() {
					defer wg.Done()
					apiTester(t, http.MethodPost, passingTest.AuthRequest, "/api/v1/register", RegisterAPI, passingTest.pass)
				}()
			}
			wg.Wait()

		}) {
			return
		}

		for _, passingTest := range registerTest2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				apiTester(t, http.MethodPost, passingTest.AuthRequest, "/api/v1/register", RegisterAPI, passingTest.pass)
			}()
		}
		wg.Wait()
	})
}
