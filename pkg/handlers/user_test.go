package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"reddit/pkg/session"
	"reddit/pkg/user"
)

const defaultReqBody = `{"username":"some_username", "password":"brevhbehvbe"}`

func TestUserHandlerLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctrlSess := gomock.NewController(t)
	defer ctrlSess.Finish()

	testRepo := user.NewMockUserRepo(ctrl)
	testSessManager := session.NewMockSessManager(ctrl)
	testHandler := &UserHandler{
		Logger:         zap.NewNop().Sugar(),
		UserRepo:       testRepo,
		SessionManager: testSessManager,
	}

	//  ошибка в чтении тела запроса
	request := httptest.NewRequest(http.MethodPost, "/api/login", &errorReader{})
	respWriter := httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status %d, got status %d", http.StatusBadRequest, resp.StatusCode)
	}

	//  запрос в плохом формате
	reqBody := `{"login"`
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
	}

	//  запрос не прошел валидацию
	reqBody = `{"username":""}`
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 422 {
		t.Errorf("expected status %d, got status %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}

	//  запрос успешный, но юзера не существует
	reqBody = `{"username":"qqqqqqqq", "password":"qqqqqqqq"}`
	testRepo.EXPECT().Login("qqqqqqqq", "qqqqqqqq").Return(nil, user.ErrNoUser)
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected status %d, got status %d", http.StatusUnauthorized, resp.StatusCode)
	}

	//  неверный пароль
	reqBody = `{"username":"qqqqqqqq", "password":"wrong_password"}`
	testRepo.EXPECT().Login("qqqqqqqq", "wrong_password").Return(nil, user.ErrBadPass)
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected status %d, got status %d", http.StatusUnauthorized, resp.StatusCode)
	}

	//  какая то ошибка сервера
	reqBody = `{"username":"qqqqqqqq", "password":"brevhbehvbe"}`
	testRepo.EXPECT().Login("qqqqqqqq", "brevhbehvbe").Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
	}

	//  не получается создать сессию
	loggedInUser := &user.User{
		ID:       "some_id",
		Username: "some_username",
	}
	testRepo.EXPECT().Login("some_username", "brevhbehvbe").Return(loggedInUser, nil)
	testSessManager.EXPECT().CreateNewSession(loggedInUser).Return("", fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(defaultReqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
	}

	//  возвращает нормально структуру с токеном
	testRepo.EXPECT().Login("some_username", "brevhbehvbe").Return(loggedInUser, nil)
	testSessManager.EXPECT().CreateNewSession(loggedInUser).Return("some_token", nil)
	request = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(defaultReqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Login(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
	}

}

func TestUserHandlerRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctrlSess := gomock.NewController(t)
	defer ctrlSess.Finish()

	testRepo := user.NewMockUserRepo(ctrl)
	testSessManager := session.NewMockSessManager(ctrl)
	testHandler := &UserHandler{
		Logger:         zap.NewNop().Sugar(),
		UserRepo:       testRepo,
		SessionManager: testSessManager,
	}

	//  ошибка в чтении тела запроса
	request := httptest.NewRequest(http.MethodPost, "/api/register", &errorReader{})
	respWriter := httptest.NewRecorder()
	testHandler.Register(respWriter, request)
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status %d, got status %d", http.StatusBadRequest, resp.StatusCode)
	}

	//  такой юзер уже существует
	reqBody := `{"username":"already_exist_username", "password":"password"}`
	testRepo.EXPECT().Register("already_exist_username", "password").Return(nil, user.ErrAlreadyExist)
	request = httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Register(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 422 {
		t.Errorf("expected status %d, got status %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}

	// какая то ошибка сервера
	reqBody = `{"username":"username", "password":"password"}`
	testRepo.EXPECT().Register("username", "password").Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(reqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Register(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
	}

	//  успешная регистрация
	registredUser := &user.User{
		ID:       "some_id",
		Username: "some_username",
	}
	testRepo.EXPECT().Register("some_username", "brevhbehvbe").Return(registredUser, nil)
	testSessManager.EXPECT().CreateNewSession(registredUser).Return("some_token", nil)
	request = httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(defaultReqBody))
	respWriter = httptest.NewRecorder()
	testHandler.Register(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
	}

}
