package data

import (
	"database/sql"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id                int64
	Email             sql.NullString
	PasswordEncrypted sql.NullString
	Name              sql.NullString
	Role              sql.NullString
}

func IsUserExistByEmail(e string) (bool, error) {
	var id int64
	err := Db.QueryRow("select id from User where email = ?", e).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

func CreateUserByEmail(e string, pe string, n string) (int64, error) {

	result, err := Db.Exec("insert into User (email, password_encrypted, name) values (?, ? ,?)", e, pe, n)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetUserByEmail(email string) (*User, error) {

	user := User{}
	row := Db.QueryRow("select id, email, password_encrypted, name, role from User where email = ?", email)
	err := row.Scan(&user.Id, &user.Email, &user.PasswordEncrypted, &user.Name, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

func CheckPassword(password string, passwordEncrypted string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordEncrypted), []byte(password))
	if err != nil {
		return false
	}
	return true
}

func EncryptPassword(password string) (string, error) {

	bp, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Errorf("EncryptPassword|ERROR|%v", err)
	}
	return string(bp), err
}
