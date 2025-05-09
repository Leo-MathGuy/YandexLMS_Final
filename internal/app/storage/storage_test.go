package storage

import (
	"crypto/sha256"
	"database/sql"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/util"
	"github.com/golang-jwt/jwt/v5"
)

func TestStorage(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
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

		if res, err := CheckPass(db, "bob", "123"); err != nil {
			t.Errorf("Error checking password: %s", err.Error())
		} else if !res {
			t.Error("Password returned false")
		}

		if res, err := CheckPass(db, "bob", "1234"); err != nil {
			t.Errorf("Error checking password: %s", err.Error())
		} else if res {
			t.Error("Password returned true")
		}

		if res, err := CheckPass(db, "bob", ""); err != nil {
			t.Errorf("Error checking password: %s", err.Error())
		} else if res {
			t.Error("Empty pawssword returned true")
		}

		if res, err := CheckPass(db, "aaa", ""); err == nil {
			t.Error("No error checking nonexistent password")
		} else if res {
			t.Error("Empty pawssword returned true")
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

		t.Run("Token valid", func(t *testing.T) {
			if token, err := CreateToken("bob"); err != nil {
				t.Fatalf("Token creation failed: %s", err.Error())
			} else if out, err := CheckToken(db, token); err != nil {
				t.Errorf("Error checking token: %s", err.Error())
			} else if out == nil {
				t.Errorf("Token not validated")
			}
		})

		t.Run("Token expired", func(t *testing.T) {
			now := time.Now().UTC()
			expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
				jwt.MapClaims{
					"user": strings.ToLower("bob"),
					"exp":  now.Add(-30 * time.Minute).UnixMilli(),
					"iat":  now.UnixMilli(),
				})

			expired, err := expiredToken.SignedString(secretKey)
			if err != nil {
				t.Fatalf("Creation of expired token failed: %s", err.Error())
			}

			u, err := CheckToken(db, expired)

			if !(u == nil && err == nil) {
				t.Errorf("Expired check failed")
			}
		})

		t.Run("Token invalid", func(t *testing.T) {
			now := time.Now().UTC()
			expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
				jwt.MapClaims{
					"user": strings.ToLower("meow"),
					"exp":  now.Add(30 * time.Minute).UnixMilli(),
					"iat":  now.UnixMilli(),
				})

			expired, err := expiredToken.SignedString(secretKey)
			if err != nil {
				t.Fatalf("Creation of expired token failed: %s", err.Error())
			}

			u, err := CheckToken(db, expired)

			if !(u == nil && err != nil) {
				t.Errorf("Invalid check failed")
			}
		})

		t.Run("Token wrong", func(t *testing.T) {
			u, err := CheckToken(db, "test")

			if !(u == nil && err != nil) {
				t.Errorf("Wrong token check failed")
			}
		})
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

		if _, err := AddExpression(&e, db, bob.ID, "2+%"); err == nil {
			t.Fatalf("No error loading invalid expression")
		}

		e.E[1].Result = 2.0
		e.E[1].Finished = true
		if *GetExpressionResult(&e, 1) != 2.0 {
			t.Errorf("Result not given")
		}
	})
}

func TestDb(t *testing.T) {
	util.Leave()
	os.Remove("testdb.db")
	defer os.Remove("testdb.db")
	t.Setenv("APPDB", "./testdb.db")
	defer t.Setenv("APPDB", "")
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Failed to connect to db: %s", r)
		}
	}()
	defer DisconnectDB()
	stop := ConnectDB()
	defer close(stop)
}

func TestTasks(t *testing.T) {
	e := Expressions{make(map[uint]*Expression), sync.RWMutex{}}

	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	err := db.Ping()
	if err != nil {
		t.Fatalf("DB Failed womp womp")
	}
	CreateTables(db)

	tasks := Tasks{make(map[uint]*Task), make([]uint, 0), 0, sync.RWMutex{}}

	if AddUser(db, "bob", "123") != nil {
		t.Fatal("Error adding user")
	}

	bob, err := GetUser(db, "bob")
	if err != nil {
		t.Fatalf("Error getting bob")
	}

	var expr *Expression
	if i, err := AddExpression(&e, db, bob.ID, "2+2"); err != nil {
		t.Fatalf("Error adding task: %s", err.Error())
	} else {
		expr = e.E[i]
	}

	if err := GenTasks(&tasks, expr); err != nil {
		t.Fatalf("Error making tasks: %s", err.Error())
	}

	if len(tasks.T) != 1 {
		t.Fatalf("Len tasks not 1")
	}

	if expr.TaskID == 0 {
		t.Fatalf("Task ID not added")
	}
	rootTask := tasks.T[expr.TaskID]

	if rootTask.LeftT != nil || rootTask.RightT != nil {
		t.Fatalf("Root task not flattened")
	}

	if rootTask.Op == nil {
		t.Fatalf("No operator")
	}

	if rootTask.Value {
		t.Fatalf("Root task value")
	}

	if rootTask.Left == nil || rootTask.Right == nil {
		t.Fatalf("Root task not flattened 2")
	}

	if *rootTask.Left != 2 || *rootTask.Right != 2 {
		t.Fatalf("Child tasks not 2")
	}

	complex := "((3.14 * (-5.2 + 7.8)) / (2.5 - (4.1 * (9.3 / -2.7)))) + (6.9 * ((1.2 - 3.4) / (8.5 + (-4.6 * 2.3)))) - ((-7.1 + 5.9) * (3.3 / (1.5 - 9.7)))"
	i, err := AddExpression(&e, db, bob.ID, complex)

	if err != nil {
		t.Fatalf("Error generating expression: %s", err.Error())
	}

	if err := GenTasks(&tasks, e.E[i]); err != nil {
		t.Fatalf("Error generating task tree: %s", err.Error())
	}

	var task *Task
	if task = GetReadyTask(&tasks); task == nil {
		t.Fatalf("No task")
	} else if task != rootTask {
		t.Fatalf("Wrong task")
	}

	if err := FinishTask(&tasks, task.ID, 4); err != nil {
		t.Fatalf("Cannot finish task: %s", err.Error())
	}

	if *rootTask.Left != 4.0 {
		t.Fatalf("Task not done")
	}

	CheckExpressions(db, &e, &tasks)

	if !expr.Finished {
		t.Fatal("Expr not finished")
	}

	if expr.Result != 4.0 {
		t.Fatalf("Wrong answer")
	}
}
