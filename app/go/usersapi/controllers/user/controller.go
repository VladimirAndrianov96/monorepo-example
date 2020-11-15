package user_controller

import (
	"encoding/json"
	"errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/gofrs/uuid"
	"go-ddd-cqrs-example/usersapi/auth"
	"go-ddd-cqrs-example/usersapi/responses"
	"go-ddd-cqrs-example/usersapi/server"
	domain_errors "go-ddd-cqrs-example/domain/errors"
	"go-ddd-cqrs-example/domain/models/user"
	"io/ioutil"
	"net/http"
)

// Register new user with the given details.
func Register(server *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		registrationReq := RegistrationRequest{}
		err = json.Unmarshal(body, &registrationReq)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		err = validation.ValidateStruct(&registrationReq,
			validation.Field(&registrationReq.EmailAddress, validation.Required, is.Email),
			validation.Field(&registrationReq.Password, validation.Required, validation.Length(6,20)),
		)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		pendingUser := user.PendingUser{
			ID:           uuid.Must(uuid.NewV4()),
			EmailAddress: registrationReq.EmailAddress,
			Password:     registrationReq.Password,
		}

		userCreatedEvent, err := user.Create(*server.DB, pendingUser)
		if err != nil {
			if errors.As(err, &user.AlreadyExists{}){
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}else{
				responses.ERROR(w, http.StatusInternalServerError, nil)
				return
			}
		}

		pkUUID, err := uuid.FromString(userCreatedEvent.UserID)
		if err != nil{
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		token, err := auth.CreateJWTToken(server.SecretKey, pkUUID)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		response := RegistrationSuccessResponse{
			Token:  *token,
			UserID: userCreatedEvent.UserID,
		}

		responses.JSON(w, http.StatusCreated, response)
	}
}

// Deactivate an active user.
func Deactivate(server *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.ExtractUserID(*server, r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		activeUser, err := user.GetActive(*server.DB, userID, nil)
		if err != nil {
			if errors.As(err, &user.IsInactive{}){
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}else{
				responses.ERROR(w, http.StatusInternalServerError, errors.New("Incorrect details"))
				return
			}
		}

		_, err = user.Deactivate(*server.DB, *activeUser)
		if err != nil {
			if errors.As(err, &domain_errors.StateConflict{}){
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}else{
				responses.ERROR(w, http.StatusInternalServerError, errors.New("Incorrect details"))
				return
			}
		}

		responses.JSON(w, http.StatusOK, StatusResponse{"User deactivated"})
	}
}

// Activate an inactive user.
func Activate(server *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.ExtractUserID(*server, r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		inactiveUser, err := user.GetInactive(*server.DB, userID, nil)
		if err != nil {
			if errors.As(err, &user.IsActive{}){
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}else{
				responses.ERROR(w, http.StatusInternalServerError, errors.New("Incorrect details"))
				return
			}
		}

		_, err = user.Activate(*server.DB, *inactiveUser)
		if err != nil {
			if errors.As(err, &domain_errors.StateConflict{}){
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}else{
				responses.ERROR(w, http.StatusInternalServerError, errors.New("Incorrect details"))
				return
			}
		}

		responses.JSON(w, http.StatusOK, StatusResponse{"User activated"})
	}
}