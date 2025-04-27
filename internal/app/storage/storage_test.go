package storage

import (
	"crypto/sha256"
	"testing"
)

func TestUsers(t *testing.T) {
	users := Users{}
	users.users = make(map[string]*User)

	users.AddUser("Bob", "123")

	if bob, found := users.users["bob"]; !found {
		t.Fatalf("bob not added")
	} else if bob.passHash != sha256.Sum256([]byte("123")) {
		t.Errorf("wrong hash")
	}

	if !users.UserExists("bob") {
		t.Fatalf("user check returned false")
	}
	if !users.UserExists("Bob") {
		t.Errorf("user check returned false")
	}
	if users.UserExists("alex") {
		t.Errorf("user check returned true")
	}

	if !users.CheckPass("bob", "123") {
		t.Errorf("password check returned false")
	}
	if !users.CheckPass("Bob", "123") {
		t.Errorf("password check returned false")
	}
	if users.CheckPass("bob", "1234") {
		t.Errorf("password check returned true")
	}
}
