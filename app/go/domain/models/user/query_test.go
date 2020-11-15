package user_test

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go-ddd-cqrs-example/api/cmd/config"
	"go-ddd-cqrs-example/api/utils"
	domain_errors "go-ddd-cqrs-example/domain/errors"
	"go-ddd-cqrs-example/domain/models/user"
	"os"
	"path"
	"runtime"
)

var _ = Describe("User loading", func() {
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

	BeforeEach(func() {
		db = conn.Begin()
	})

	AfterEach(func() {
		_ = db.Rollback()
	})

	Describe("Getting active users", func() {
		var userID uuid.UUID

		When("user is not active", func() {
			BeforeEach(func() {
				userID = uuid.Must(uuid.NewV4())

				err := db.Create(&user.User{
					ID:           userID,
					EmailAddress: "user@example.com",
					Password:     "someHashedPassword",
					IsActive:     false,
					Version:      1,
				}).Error

				Expect(err).To(BeNil())
			})

			Specify("a user inactive error is returned", func() {
				activeUser, err := user.GetActive(*db, userID, nil)

				Expect(activeUser).To(BeNil())
				Expect(errors.As(err, &user.IsInactive{})).To(BeTrue())
			})
		})

		When("an invalid version tag is specified", func() {
			BeforeEach(func() {
				userID = uuid.Must(uuid.NewV4())

				err := db.Create(&user.User{
					ID:           userID,
					EmailAddress: "user@example.com",
					Password:     "someHashedPassword",
					IsActive:     true,
					Version:      1,
				}).Error

				Expect(err).To(BeNil())
			})

			Specify("an invalid version error is returned", func() {
				v := uint32(3)
				activeUser, err := user.GetActive(*db, userID, &v)

				Expect(activeUser).To(BeNil())
				Expect(errors.As(err, &domain_errors.InvalidVersion{})).To(BeTrue())
			})
		})
	})

	Describe("Getting inactive users", func() {
		var userID uuid.UUID

		When("user is active", func() {
			BeforeEach(func() {
				userID = uuid.Must(uuid.NewV4())

				err := db.Create(&user.User{
					ID:           userID,
					EmailAddress: "user@example.com",
					Password:     "someHashedPassword",
					IsActive:     true,
					Version:      1,
				}).Error

				Expect(err).To(BeNil())
			})

			Specify("a user active error is returned", func() {
				inactiveUser, err := user.GetInactive(*db, userID, nil)

				Expect(inactiveUser).To(BeNil())
				Expect(errors.As(err, &user.IsActive{})).To(BeTrue())
			})
		})

		When("an invalid version tag is specified", func() {
			BeforeEach(func() {
				userID = uuid.Must(uuid.NewV4())

				err := db.Create(&user.User{
					ID:           userID,
					EmailAddress: "user@example.com",
					Password:     "someHashedPassword",
					IsActive:     false,
					Version:      1,
				}).Error

				Expect(err).To(BeNil())
			})

			Specify("an invalid version error is returned", func() {
				v := uint32(3)
				inactiveUser, err := user.GetInactive(*db, userID, &v)

				Expect(inactiveUser).To(BeNil())
				Expect(errors.As(err, &domain_errors.InvalidVersion{})).To(BeTrue())
			})
		})
	})
})
