package storage

import (
	"crypto/sha256"
	"database/sql"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/util"
)

func TestStorage(t *testing.T) {
	util.Leave()
	os.Remove("test.db")
	defer os.Remove("test.db")

	db, _ := sql.Open("sqlite3", "./test.db")
	err := db.Ping()
	if err != nil {
		t.Fatalf("DB Failed womp womp")
	}
	CreateTables(db)

	proceed := t.Run("User", func(t *testing.T) {
		if err := AddUser(db, "bob", "123"); err != nil {
			t.Fatal(err.Error())
		}

		if n, _ := UserExists(db, "bob"); !n {
			t.Fatal("bob not found")
		}
		if n, _ := UserExists(db, "Bob"); !n {
			t.Error("Bob not found")
		}
		if n, _ := UserExists(db, "tom"); n {
			t.Fatal("tom was found")
		}

		user1, err1 := GetUser(db, "bob")
		user2, err2 := GetUser(db, "Bob")
		_, err3 := GetUser(db, "alice")

		if user1 == nil || user2 == nil || err1 != nil || err2 != nil {
			t.Fatalf("bob was not gotten")
		}

		if !reflect.DeepEqual(*user1, *user2) {
			t.Error("bob =/= Bob :(")
		}

		h := sha256.Sum256([]byte("123"))
		bobExpect := &User{1, "bob", h[:]}
		if !reflect.DeepEqual(*user1, *bobExpect) {
			t.Error("bob isn't expected bob")
		}

		if err3 == nil {
			t.Error("alice was found")
		}

		AddUser(db, "eVe", "456")
		users, err := GetUsers(db)
		if err != nil {
			t.Error(err.Error())
		} else {
			if len(users) != 2 {
				t.Error("Not 2 users")
			} else {
				if users[0].Login != "bob" {
					t.Error("WHERE IS BOB")
				}
				if users[1].Login != "eve" {
					t.Error("No eve")
				}
			}
		}

		if token, err := CreateToken("bob"); err != nil {
			t.Fatalf("Token creation failed: %s", err.Error())
		} else if out, err := CheckToken(db, token); err != nil {
			t.Errorf("Error checking token: %s", err.Error())
		} else if out == nil {
			t.Errorf("Token not validated")
		}

	})

	if !proceed {
		return
	}

	t.Run("Expression", func(t *testing.T) {
		e := Expressions{make(map[uint]*Expression), sync.RWMutex{}}

		bob, err := GetUser(db, "bob")
		if err != nil {
			t.Fatalf("Cannot get bob: %s", err.Error())
		}
		if _, err := AddExpression(&e, db, bob.ID, "2+2"); err != nil {
			t.Fatalf("Cannot add expr: %s", err.Error())
		}

		if len(e.E) < 1 {
			t.Fatalf("Nothing was added")
		}

		ex := e.E[1]

		if ex.UID != bob.ID {
			t.Error("Wrong UID")
		}
		if ex.Gen.Left == nil || ex.Gen.Right == nil || *ex.Gen.Left.Value != 2.0 || *ex.Gen.Right.Value != 2.0 {
			t.Error("Wrong values")
		}
		if ex.Gen.Op == nil || *ex.Gen.Op != rune('+') {
			t.Error("Wrong op")
		}

		if ex.Finished {
			t.Error("Finished, but not expected to be")
		}

		if ex.Result != 0.0 {
			t.Error("Result non zero")
		}

		e = Expressions{make(map[uint]*Expression), sync.RWMutex{}}
		if err := LoadExpressions(db, &e); err != nil {
			t.Fatalf("Error loading: %s", err.Error())
		}

		if len(e.E) < 1 {
			t.Fatalf("Nothing was added 2")
		}

		ex2 := e.E[1]

		if !reflect.DeepEqual(*ex, *ex2) {
			t.Errorf("Not loaded correctly")
		}
	})
}
