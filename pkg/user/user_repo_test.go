package user

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-sql-driver/mysql"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reddit/pkg/idgenerator"
)

func TestLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("can not create mock")
	}
	defer db.Close()
	dbRepo := &UserDBRepo{
		DB: db,
	}
	idGen := &idgenerator.TestIDGenerator{}
	repo := NewUserMemoryRepository(dbRepo, idGen)

	username := "some_username"
	password := "1464acd6765f91fccd3f5bf4f14ebb7ca69f53af91b0a5790c2bba9d8819417b" // захешированный "some_password"
	id := "some_id"

	// какая то ошибка базы данных
	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs(username).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.Login(username, password)
	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// юзера не существует
	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Login(username, password)
	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// неверный пароль
	rows := sqlmock.NewRows([]string{"id", "username", "password"})
	expect := []*User{
		{
			ID:       id,
			Username: username,
			password: password,
		},
	}
	for _, currentUser := range expect {
		rows = rows.AddRow(currentUser.ID, currentUser.Username, currentUser.password)
	}
	wrongPassword := "wrong_password"
	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs(username).
		WillReturnRows(rows)
	loggedInUser, err := repo.Login(username, wrongPassword)
	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if !errors.Is(err, ErrBadPass) {
		t.Errorf("expected error %s, got error %s", ErrBadPass, err)
		return
	}
	if loggedInUser != nil {
		t.Errorf("unexpected non nil user")
		return
	}

	// успешная авторизация
	expectedUser := &User{
		ID:       id,
		Username: username,
		password: "1464acd6765f91fccd3f5bf4f14ebb7ca69f53af91b0a5790c2bba9d8819417b",
	}
	rows = sqlmock.NewRows([]string{"id", "username", "password"})
	for _, currentUser := range expect {
		rows = rows.AddRow(currentUser.ID, currentUser.Username, currentUser.password)
	}
	unHashedPassword := "some_password"
	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE").
		WithArgs(username).
		WillReturnRows(rows)
	loggedInUser, err = repo.Login(username, unHashedPassword)
	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if !reflect.DeepEqual(loggedInUser, expectedUser) {
		t.Errorf("wrong user: expected %v, got %v", expectedUser, loggedInUser)
		return
	}

}

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("can not create mock")
	}
	defer db.Close()
	dbRepo := &UserDBRepo{
		DB: db,
	}
	idGen := &idgenerator.TestIDGenerator{}
	repo := NewUserMemoryRepository(dbRepo, idGen)

	userID := idGen.GenerateID(16)

	userToInsert := &User{
		ID:       userID,
		Username: "some_username",
		password: "1464acd6765f91fccd3f5bf4f14ebb7ca69f53af91b0a5790c2bba9d8819417b",
	}

	// какая то ошибка бд

	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(userToInsert.ID, userToInsert.Username, userToInsert.password).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.Register(userToInsert.Username, "some_password")

	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	// юзер уже существует
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(userToInsert.ID, userToInsert.Username, userToInsert.password).
		WillReturnError(&mysql.MySQLError{
			Number: 1062,
		})

	_, err = repo.Register(userToInsert.Username, "some_password")

	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if !errors.Is(err, ErrAlreadyExist) {
		t.Errorf("expected error %s, got error %s", ErrAlreadyExist, err)
		return
	}

	// успешная регистрация
	mock.
		ExpectExec(`INSERT INTO users`).
		WithArgs(userToInsert.ID, userToInsert.Username, userToInsert.password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	registredUser, err := repo.Register(userToInsert.Username, "some_password")

	if err := mock.ExpectationsWereMet(); err != nil { // nolint govet
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if !reflect.DeepEqual(registredUser, userToInsert) {
		t.Errorf("wrong result: expected: %v, got %v", userToInsert, registredUser)
		return
	}

}
