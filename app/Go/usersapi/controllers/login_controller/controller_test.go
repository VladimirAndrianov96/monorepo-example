package login_controller_test

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go-ddd-cqrs-example/domain/models/user"
	"go-ddd-cqrs-example/usersapi/cmd/config"
	"go-ddd-cqrs-example/usersapi/controllers/login_controller"
	"go-ddd-cqrs-example/usersapi/routes"
	"go-ddd-cqrs-example/usersapi/server"
	"go-ddd-cqrs-example/usersapi/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
)

var _ = Describe("Login controller", func() {
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
	viper.AddConfigPath(dir + "/usersapi/cmd/config")
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

	Describe("Logging into new account", func() {
		var UserID uuid.UUID
		var usr user.PendingUser
		var usr2 user.PendingUser

		BeforeEach(func() {
			UserID = uuid.Must(uuid.NewV4())
			usr = user.PendingUser{
				ID:           UserID,
				EmailAddress: "user@example.com",
				Password:     "password",
			}

			_, err := user.Create(*db, usr)
			Expect(err).To(BeNil())

			User2ID := uuid.Must(uuid.NewV4())
			usr2 = user.PendingUser{
				ID:           User2ID,
				EmailAddress: "user2@example.com",
				Password:     "password",
			}

			_, err = user.Create(*db, usr2)
			Expect(err).To(BeNil())

			usr2Active, err := user.GetActive(*db, usr2.ID, nil)
			Expect(err).To(BeNil())

			_, err = user.Deactivate(*db, *usr2Active)
			Expect(err).To(BeNil())
		})

		When("Login request is sent", func() {
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
						statusCode:   http.StatusOK,
						errorMessage: "",
					},
					{
						email:        usr.EmailAddress,
						password:     "WrongPassword",
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "Incorrect details",
					},
					{
						email:        "Wrongemail@mail.com",
						password:     usr.Password,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "Incorrect details",
					},
					{
						email:        "Wrong email",
						password:     usr.Password,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "email_address: must be a valid email address.",
					},
					{
						email:        "",
						password:     usr.Password,
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "email_address: cannot be blank.",
					},
					{
						email:        usr.EmailAddress,
						password:     "",
						statusCode:   http.StatusUnprocessableEntity,
						errorMessage: "password: cannot be blank.",
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
						errorMessage: "Invariant failed: User is inactive",
					},
				}

				for _, v := range samples {
					loginRequest := login_controller.LoginRequest{
						EmailAddress: v.email,
						Password:     v.password,
					}

					requestBody, err := json.Marshal(loginRequest)
					Expect(err).To(gomega.BeNil())

					req, err := http.NewRequest("POST", "/usersapi/login", bytes.NewBufferString(string(requestBody)))
					Expect(err).To(gomega.BeNil())

					rr := httptest.NewRecorder()
					handler := http.HandlerFunc(login_controller.Login(&srv))
					handler.ServeHTTP(rr, req)

					responseMap := make(map[string]interface{})

					err = json.Unmarshal([]byte(rr.Body.String()), &responseMap)
					Expect(err).To(gomega.BeNil())

					Expect(rr.Code).To(Equal(v.statusCode))

					if v.statusCode == 200 {
						Expect(responseMap["user_id"]).ToNot(Equal(""))
						Expect(responseMap["token"]).ToNot(Equal(""))
					}

					if v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
						Expect(responseMap["error"]).To(Equal(v.errorMessage))
					}
				}
			})
		})
	})
})
