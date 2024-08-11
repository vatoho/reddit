package user

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type UserDBRepo struct {
	DB *sql.DB
}

func (u *UserDBRepo) FindUserByUsernameDB(username string) (*User, error) {
	loginUser := &User{}
	err := u.DB.
		QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).
		Scan(&loginUser.ID, &loginUser.Username, &loginUser.password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoUser
		}
		return nil, err
	}
	return loginUser, nil
}

func (u *UserDBRepo) AddNewUserDB(newUser *User) error {
	_, err := u.DB.Exec(
		"INSERT INTO users (`id`, `username`, `password`) VALUES (?, ?, ?)",
		newUser.ID,
		newUser.Username,
		newUser.password,
	)
	if err != nil {
		if isAlreadyExists(err) {
			return ErrAlreadyExist
		}
		return err
	}
	return nil
}

func isAlreadyExists(err error) bool {
	mysqlError, ok := err.(*mysql.MySQLError)
	return ok && mysqlError.Number == 1062
}
