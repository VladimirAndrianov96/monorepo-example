package user_controller_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go-ddd-cqrs-example/api/auth"
	"go-ddd-cqrs-example/api/cmd/config"
	user_controller "go-ddd-cqrs-example/api/controllers/user"
	"go-ddd-cqrs-example/api/routes"
	"go-ddd-cqrs-example/api/server"
	"go-ddd-cqrs-example/api/utils"
	"go-ddd-cqrs-example/domain/models/user"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
)

var _ = Describe("User controller", func() {
	var (
		db *gorm.DB
	)

	// Hotfix, fix inconsistent current directory to get configuration file.
	// TODO find better way to handle this.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	// Set up database connection using configuration details.
	cfg := config.Config{}
	viper.AddConfigPath(dir+"/api/cmd/config")
	viper.SetConfigName("configuration")
	viper.ReadInConfig()
	viper.Unmarshal(&cfg)
	conn, err := utils.GetDB(
		cfg.DBDriver,
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBPort,
		cfg.DBHost,
		cfg.DBName,
	)
	Expect(err).To(BeNil())

	srv := server.Server{}
	srv.SecretKey = cfg.SecretKey
	srv.Router = mux.NewRouter()
	routes.InitializeRoutes(&srv)

	BeforeEach(func() {
		db = conn.Begin()
		srv.DB = db
	})

	AfterEach(func() {
		_ = db.Rollback()
	})

	Describe("Registering new user", func() {
		var UserID uuid.UUID
		var usr user.PendingUser
		var usr2 user.PendingUser

		BeforeEach(func() {
			UserID = uuid.Must(uuid.NewV4())
			usr = user.PendingUser{
				ID:           UserID,
				EmailAddress: "user@example.com",
				Password: "password",
			}

			User2ID := uuid.Must(uuid.NewV4())
			usr2 = user.PendingUser{
				ID:           User2ID,
				EmailAddress: "user2@example.com",
				Password: "password",
			}

			_, err = user.Create(*db, usr2)
			Expect(err).To(BeNil())
		})

		When("Registration request is sent", func() {
			Specify("The response returned", func() {
				samples := []struct {
					email        string
					password     string
					statusCode   int
					errorMessage string
				}{
					{
						email:        usr.EmailAddress,
						password:     usr.Password, // Non-hashed password is required to sign in.
						statusCode:   http.StatusCreated,
						errorMessage: "",
					},
					{
						email:        "Wrong email",
						password:     usr.Password,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "email_address: must be a valid email address.",
					},
					{
						email:        "unique@example.com",
						password:     "",
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "password: cannot be blank.",
					},
					{
						email:        "",
						password:     usr.Password,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "email_address: cannot be blank.",
					},
					{
						email:        "Wrong email",
						password:     "",
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "email_address: must be a valid email address; password: cannot be blank.",
					},
					{
						email:        usr2.EmailAddress,
						password:     usr2.Password, // Non-hashed password is required to sign in.
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "User already exists",
					},
				}

				for _, v := range samples {
					registrationRequest := user_controller.RegistrationRequest{
						EmailAddress: v.email,
						Password: v.password,
					}

					requestBody, err := json.Marshal(registrationRequest)
					Expect(err).To(gomega.BeNil())

					req, err := http.NewRequest("POST", "/api/register", bytes.NewBufferString(string(requestBody)))
					Expect(err).To(gomega.BeNil())

					rr := httptest.NewRecorder()
					handler := http.HandlerFunc(user_controller.Register(&srv))
					handler.ServeHTTP(rr, req)

					responseMap := make(map[string]interface{})

					err = json.Unmarshal([]byte(rr.Body.String()), &responseMap)
					Expect(err).To(gomega.BeNil())

					Expect(rr.Code).To(Equal(v.statusCode))

					if v.statusCode == http.StatusOK {
						Expect(responseMap["user_id"]).ToNot(Equal(""))
						Expect(responseMap["token"]).ToNot(Equal(""))

						usrFetched, err := user.GetActiveByEmail(*srv.DB, usr.EmailAddress, nil)
						Expect(usrFetched).ToNot(BeNil())
						Expect(err).To(BeNil())
					}

					if v.statusCode == http.StatusUnprocessableEntity || v.statusCode == http.StatusInternalServerError && v.errorMessage != "" {
						Expect(responseMap["error"]).To(Equal(v.errorMessage))
					}
				}
			})
		})
	})

	Describe("Deactivaing a user", func() {
		var UserID uuid.UUID
		var usr user.PendingUser
		var tokenString string

		BeforeEach(func() {
			UserID = uuid.Must(uuid.NewV4())
			usr = user.PendingUser{
				ID:           UserID,
				EmailAddress: "user@example.com",
				Password: "password",
			}

			_, err = user.Create(*db, usr)
			Expect(err).To(BeNil())

			//Log in the user and get the authentication token.
			token, _, err := auth.SignIn(&srv, usr.EmailAddress, "password")
			Expect(err).To(gomega.BeNil())

			tokenString = fmt.Sprintf("Bearer %v", *token)
		})

		When("Deactivation request is sent", func() {
			Specify("The response returned", func() {
				samples := []struct {
					token        string
					statusCode   int
					errorMessage string
				}{
					{
						token: tokenString,
						statusCode:   http.StatusOK,
						errorMessage: "",
					},
					{
						token: tokenString,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "Invariant failed: User is inactive",
					},
					{
						token: "wrongToken",
						statusCode:   http.StatusUnauthorized,
						errorMessage: "",
					},
				}

				for _, v := range samples {
					req, err := http.NewRequest("POST", "/api/deactivate/current", bytes.NewBufferString(""))
					Expect(err).To(gomega.BeNil())

					rr := httptest.NewRecorder()
					handler := http.HandlerFunc(user_controller.Deactivate(&srv))
					req.Header.Set("Authorization", v.token)
					handler.ServeHTTP(rr, req)

					responseMap := make(map[string]interface{})
					err = json.Unmarshal([]byte(rr.Body.String()), &responseMap)
					Expect(err).To(gomega.BeNil())

					Expect(rr.Code).To(Equal(v.statusCode))
					if v.statusCode == http.StatusOK {
						Expect(responseMap["response"]).To(Equal("User deactivated"))
					}

					if v.statusCode == http.StatusUnprocessableEntity || v.statusCode == http.StatusInternalServerError && v.errorMessage != "" {
						Expect(responseMap["error"]).To(Equal(v.errorMessage))
					}
				}
			})
		})
	})

	Describe("Activaing a user", func() {
		var UserID uuid.UUID
		var usr user.PendingUser
		var tokenString string

		BeforeEach(func() {
			UserID = uuid.Must(uuid.NewV4())
			usr = user.PendingUser{
				ID:           UserID,
				EmailAddress: "user@example.com",
				Password: "password",
			}

			_, err = user.Create(*db, usr)
			Expect(err).To(BeNil())

			activeUser, err := user.GetActive(*db, usr.ID, nil)
			Expect(err).To(BeNil())

			//Log in the user and get the authentication token.
			token, _, err := auth.SignIn(&srv, usr.EmailAddress, "password")
			Expect(err).To(gomega.BeNil())

			_, err = user.Deactivate(*db, *activeUser)
			Expect(err).To(BeNil())

			tokenString = fmt.Sprintf("Bearer %v", *token)
		})

		When("Activation request is sent", func() {
			Specify("The response returned", func() {
				samples := []struct {
					token        string
					statusCode   int
					errorMessage string
				}{
					{
						token: tokenString,
						statusCode:   http.StatusOK,
						errorMessage: "",
					},
					{
						token: tokenString,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "Invariant failed: User is active",
					},
					{
						token: "wrongToken",
						statusCode:   http.StatusUnauthorized,
						errorMessage: "",
					},
				}

				for _, v := range samples {
					req, err := http.NewRequest("POST", "/api/activate/current", bytes.NewBufferString(""))
					Expect(err).To(gomega.BeNil())

					rr := httptest.NewRecorder()
					handler := http.HandlerFunc(user_controller.Activate(&srv))
					req.Header.Set("Authorization", v.token)
					handler.ServeHTTP(rr, req)

					responseMap := make(map[string]interface{})
					err = json.Unmarshal([]byte(rr.Body.String()), &responseMap)
					Expect(err).To(gomega.BeNil())

					Expect(rr.Code).To(Equal(v.statusCode))
					if v.statusCode == http.StatusOK {
						Expect(responseMap["response"]).To(Equal("User activated"))
					}

					if v.statusCode == http.StatusUnprocessableEntity || v.statusCode == http.StatusInternalServerError && v.errorMessage != "" {
						Expect(responseMap["error"]).To(Equal(v.errorMessage))
					}
				}
			})
		})
	})
})
