package storage

import (
	"crypto/sha256"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MARK: Users
type User struct {
	login    string
	passHash [32]byte
}

type Users struct {
	users map[string]*User
	sync.RWMutex
}

var U Users

func InitUsers() {
	U = Users{}
	U.users = make(map[string]*User)
}

func (u *Users) AddUser(login, password string) {
	u.Lock()
	defer u.Unlock()
	p := &User{login, sha256.Sum256([]byte(password))}
	u.users[strings.ToLower(login)] = p
}

func (u *Users) UserExists(login string) (exists bool) {
	u.RLock()
	defer u.RUnlock()

	_, exists = u.users[strings.ToLower(login)]
	return exists
}

func (u *Users) CheckPass(login, password string) (correct bool) {
	u.RLock()
	defer u.RUnlock()

	return reflect.DeepEqual(sha256.Sum256([]byte(password)), u.users[strings.ToLower(login)].passHash)
}

// MARK: JWT
var secretKey = []byte("ultra-secret-key")

func CreateToken(username string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user": strings.ToLower(username),
			"nbf":  now.Add(time.Minute).Unix(),
			"exp":  now.Add(30 * time.Minute).Unix(),
			"iat":  now.Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
