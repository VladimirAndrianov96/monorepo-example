package login_controller

import (
	"encoding/json"
	"errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"go-ddd-cqrs-example/domain/models/user"
	"go-ddd-cqrs-example/usersapi/auth"
	"go-ddd-cqrs-example/usersapi/responses"
	"go-ddd-cqrs-example/usersapi/server"
	"io/ioutil"
	"net/http"
)

// Login handles the authentication.
func Login(server *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		loginReq := LoginRequest{}
		err = json.Unmarshal(body, &loginReq)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		err = validation.ValidateStruct(&loginReq,
			validation.Field(&loginReq.EmailAddress, validation.Required, is.Email),
			validation.Field(&loginReq.Password, validation.Required, validation.Length(6, 20)),
		)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		token, userID, err := auth.SignIn(server, loginReq.EmailAddress, loginReq.Password)
		if err != nil {
			if errors.As(err, &user.IsInactive{}) {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			} else {
				responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("Incorrect details"))
				return
			}
		}

		response := loginResponse{
			Token:  *token,
			UserID: *userID,
		}

		responses.JSON(w, http.StatusOK, response)
	}
}
