package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"go.uber.org/zap"
	"reddit/pkg/response"

	"reddit/pkg/session"
	"reddit/pkg/user"
)

type UserHandler struct {
	UserRepo       user.UserRepo
	SessionManager session.SessManager
	Logger         *zap.SugaredLogger
}
type LoginRegisterRequestBody struct {
	Password string `json:"password" valid:"required,length(8|255)"`
	Username string `json:"username" valid:"required,matches(^[a-zA-Z0-9_]+$)"`
}

func (u *LoginRegisterRequestBody) Validate() []string {
	_, err := govalidator.ValidateStruct(u)
	validationErrors := make([]string, 0)
	if err == nil {
		return validationErrors
	}
	if allErrs, ok := err.(govalidator.Errors); ok {
		for _, fld := range allErrs {
			validationErrors = append(validationErrors, fld.Error())
		}
	}
	return validationErrors
}

type loginRegisterResponseBody struct {
	Token string `json:"token"`
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	userFromLoginForm, err := checkRequestFormat(uh.Logger, w, r)
	if err != nil || userFromLoginForm == nil {
		errText := fmt.Sprintf(`{"message": "error in login request: %s"}`, err)
		response.WriteResponse(uh.Logger, w, []byte(errText), http.StatusUnauthorized)
		return
	}
	loggedInUser, err := uh.UserRepo.Login(userFromLoginForm.Username, userFromLoginForm.Password)

	if errors.Is(err, user.ErrNoUser) {
		response.WriteResponse(uh.Logger, w, []byte(`{"message": "user not found"}`), http.StatusUnauthorized)
		return
	}
	if errors.Is(err, user.ErrBadPass) {
		response.WriteResponse(uh.Logger, w, []byte(`{"message": "invalid password"}`), http.StatusUnauthorized)
		return
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in getting user by login and password: %s"}`, err)
		response.WriteResponse(uh.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	uh.HandleGetToken(w, loggedInUser)

}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	userFromLoginForm, err := checkRequestFormat(uh.Logger, w, r)
	if err != nil || userFromLoginForm == nil {
		return
	}
	newUser, err := uh.UserRepo.Register(userFromLoginForm.Username, userFromLoginForm.Password)

	if errors.Is(err, user.ErrAlreadyExist) {
		response.WriteResponse(uh.Logger, w, []byte(`{"message": "user already exists"}`), http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "unknown error occured in register: %s"}`, err)
		response.WriteResponse(uh.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	uh.HandleGetToken(w, newUser)
}

func (uh *UserHandler) HandleGetToken(w http.ResponseWriter, newUser *user.User) {
	token, err := uh.SessionManager.CreateNewSession(newUser)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in session creation: %s"}`, err)
		response.WriteResponse(uh.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	resp := loginRegisterResponseBody{token}
	tokenJSON, err := json.Marshal(&resp)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding response: %s"}`, err)
		response.WriteResponse(uh.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	uh.Logger.Infof("new token: %s", token)
	response.WriteResponse(uh.Logger, w, tokenJSON, http.StatusOK)
}

func checkRequestFormat(logger *zap.SugaredLogger, w http.ResponseWriter, r *http.Request) (*LoginRegisterRequestBody, error) {
	rBody, err := io.ReadAll(r.Body)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in reading request body: %s"}`, err)
		response.WriteResponse(logger, w, []byte(errText), http.StatusBadRequest)
		return nil, err
	}
	userFromLoginForm := &LoginRegisterRequestBody{}
	err = json.Unmarshal(rBody, userFromLoginForm)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in decoding user: %s"}`, err)
		response.WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
		return nil, err
	}
	if validationErrors := userFromLoginForm.Validate(); len(validationErrors) != 0 {
		errorsJSON, err := json.Marshal(validationErrors)
		if err != nil {
			errText := fmt.Sprintf(`{"message": "error in decoding validation errors: %s"}`, err)
			response.WriteResponse(logger, w, []byte(errText), http.StatusInternalServerError)
			return nil, err
		}
		logger.Errorf("login form did not pass validation: %s", err)
		response.WriteResponse(logger, w, errorsJSON, http.StatusUnprocessableEntity)
		return nil, err
	}
	return userFromLoginForm, nil
}
