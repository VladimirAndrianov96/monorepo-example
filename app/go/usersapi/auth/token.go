package auth

import (
	"errors"
	"fmt"
	"go-ddd-cqrs-example/usersapi/server"
	"go-ddd-cqrs-example/domain/models/user"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
)

// CreateJWTToken with the configured expiration time.
func CreateJWTToken(secretKey string, uid uuid.UUID) (*string, error) {
	timeNow := time.Now()
	exp := timeNow.Add(time.Hour * 24 * 7)

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = uid
	claims["exp"] = exp
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSigned, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &tokenSigned, err
}

// CheckJWTTokenValidity to ensure that every request passes only with the valid token.
func CheckJWTTokenValidity(server server.Server, r *http.Request) (bool, error) {
	tokenString := extractJWTToken(r)
	token, err := jwt.Parse(tokenString, getTokenKeyFunc(server.SecretKey))
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid := claims["user_id"]
		_, err := uuid.FromString(fmt.Sprintf("%v", uid))
		if err != nil {
			return false, err
		}

		isAuthorized := claims["authorized"]
		if isAuthorized != true {
			return false, errors.New("Invalid token")
		}
	}

	return true, nil
}

// ExtractUserID from the valid token claims.
func ExtractUserID(server server.Server, r *http.Request) (uuid.UUID, error) {
	tokenString := extractJWTToken(r)
	token, err := jwt.Parse(tokenString, getTokenKeyFunc(server.SecretKey))
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid := claims["user_id"]
		id, err := uuid.FromString(fmt.Sprintf("%v", uid))
		if err != nil {
			return id, err
		}
		return id, nil
	}

	return uuid.Nil, nil
}

// Passed inside the jwt.Parse function which internally validates the token.
// If token signed correctly then it uses the key to verify the signature.
func getTokenKeyFunc(secretKey string)  jwt.Keyfunc {
	return func(token * jwt.Token)(interface {}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	}
}

func extractJWTToken(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}

	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}

	return ""
}

// SignIn if password is correct and return a token.
func SignIn(server *server.Server, email, password string) (*string, *string, error) {
	var err error

	userReceived, err := user.GetActiveByEmail(*server.DB, email,nil)
	if err != nil {
		if errors.As(err, &user.IsInactive{}){
			return nil, nil, err
		}
		return nil, nil, errors.New("Incorrect details")
	}

	err = user.VerifyUserPassword(server.DB, email, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, err
	}

	token, err := CreateJWTToken(server.SecretKey, userReceived.ID)
	if err != nil {
		log.Panic(err)
		return nil, nil, err
	}

	userID := userReceived.ID.String()

	return token, &userID, nil
}




