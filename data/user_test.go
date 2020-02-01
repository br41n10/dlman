package data

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestEncryptPassword(t *testing.T) {
	password := "123456"
	pe, err := EncryptPassword(password)
	if err != nil {
		t.Fail()
	}

	err = bcrypt.CompareHashAndPassword([]byte(pe), []byte(password))
	if err != nil {
		t.Fail()
	}
}

func TestIsUserExistByEmail(t *testing.T) {
	exist, err := IsUserExistByEmail("notexist@example.com")
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if exist != false {
		t.Errorf("TestIsUserExistByEmail|ERROR|exist should be false")
		t.Fail()
	}
	// fine
}

func TestCreateUserByEmail(t *testing.T) {
	email := "user@example.com"
	pe, _ := EncryptPassword("123456")

	userId, err := CreateUserByEmail(email, pe, email)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Logf("new user created: %d", userId)
}
