package user_test

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go-ddd-cqrs-example/usersapi/cmd/config"
	"go-ddd-cqrs-example/usersapi/utils"
	domain_errors "go-ddd-cqrs-example/domain/errors"
	"go-ddd-cqrs-example/domain/models/user"
	"os"
	"path"
	"runtime"
)


var _ = Describe("Managing users", func() {
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
	viper.AddConfigPath(dir+"/usersapi/cmd/config")
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

	Describe("Creating a user", func() {
		var pendingUser user.PendingUser

		BeforeEach(func() {
			pendingUser = user.PendingUser{
				ID:           uuid.Must(uuid.NewV4()),
				EmailAddress: "  User@example.com  ",
				Password:     "someHashedPassword",
			}
		})

		When("the user is created", func() {
			Specify("the returned event", func() {
				event, err := user.Create(*db, pendingUser)

				Expect(err).To(BeNil())
				Expect(event).To(Equal(&user.UserCreated{
					UserID:       pendingUser.ID.String(),
					EmailAddress: "user@example.com",
					Version:      1,
				}))
			})

			Specify("the user is persisted in the database", func() {
				_, err := user.Create(*db, pendingUser)

				Expect(err).To(BeNil())

				u := user.User{}
				err = db.Model(&u).Where("id = ?", pendingUser.ID).Take(&u).Error

				Expect(err).To(BeNil())
				Expect(u.ID).To(Equal(pendingUser.ID))
				Expect(u.EmailAddress).To(Equal("user@example.com"))
				Expect(u.IsActive).To(Equal(true))
				Expect(u.Version).To(Equal(uint32(1)))
			})
		})

		When("a user with specified email address already exists in the system", func() {
			BeforeEach(func() {
				err := db.Create(&user.User{
					ID:           uuid.Must(uuid.NewV4()),
					EmailAddress: "user@example.com",
					Password:     "someHashedPassword",
					IsActive:     true,
					Version:      1,
				}).Error

				Expect(err).To(BeNil())
			})

			Specify("a user already exists error is returned", func() {
				pendingUser := user.PendingUser{
					ID:           uuid.Must(uuid.NewV4()),
					EmailAddress: "user@example.com",
					Password:     "someHashedPassword",
				}

				event, err := user.Create(*db, pendingUser)

				Expect(event).To(BeNil())
				Expect(errors.As(err, &user.AlreadyExists{})).To(BeTrue())
			})
		})
	})

	Describe("Deactivating a user", func() {
		var UserID uuid.UUID

		BeforeEach(func() {
			UserID = uuid.Must(uuid.NewV4())

			err := db.Create(&user.User{
				ID:           UserID,
				EmailAddress: "user@example.com",
				IsActive:     true,
				Version:      1,
			}).Error
			Expect(err).To(BeNil())
		})

		When("the user is deactivated", func() {
			Specify("the returned event", func() {
				activeUser, err := user.GetActive(*db, UserID, nil)
				Expect(err).To(BeNil())

				event, err := user.Deactivate(*db, *activeUser)

				Expect(err).To(BeNil())
				Expect(event).To(Equal(&user.UserDeactivated{
					UserID:  activeUser.ID.String(),
					Version: 2,
				}))
			})

			Specify("the deactivated user is persisted in the database", func() {
				activeUser, err := user.GetActive(*db, UserID, nil)
				Expect(err).To(BeNil())

				_, err = user.Deactivate(*db, *activeUser)

				Expect(err).To(BeNil())

				u := user.User{}
				err = db.Model(&u).Where("id = ?", activeUser.ID).Take(&u).Error

				Expect(err).To(BeNil())
				Expect(u.ID).To(Equal(activeUser.ID))
				Expect(u.EmailAddress).To(Equal("user@example.com"))
				Expect(u.IsActive).To(Equal(false))
				Expect(u.Version).To(Equal(uint32(2)))
			})
		})

		When("the user's state has been modified during deactivation", func() {
			Specify("a state conflict error is returned", func() {
				activeUser, err := user.GetActive(*db, UserID, nil)
				Expect(err).To(BeNil())

				// Simulate a concurrent action on the entity by increasing its version.
				result := db.Model(&user.User{}).
					Where("id = ?", activeUser.ID).
					Updates(user.User{
						Version: activeUser.Version+1,
					},
				)

				Expect(err).To(BeNil())
				Expect(result.RowsAffected).To(Equal(int64(1)))

				inactiveUser, err := user.Deactivate(*db, *activeUser)

				Expect(errors.As(err, &domain_errors.StateConflict{})).To(BeTrue())
				Expect(inactiveUser).To(BeNil())
			})
		})
	})

	Describe("Activating a user", func() {
		var UserID uuid.UUID

		BeforeEach(func() {
			UserID = uuid.Must(uuid.NewV4())

			err := db.Create(&user.User{
				ID:           UserID,
				EmailAddress: "user@example.com",
				IsActive:     false,
				Version:      1,
			}).Error
			Expect(err).To(BeNil())
		})

		When("the user is activated", func() {
			Specify("the returned event", func() {
				inactiveUser, err := user.GetInactive(*db, UserID, nil)
				Expect(err).To(BeNil())

				event, err := user.Activate(*db, *inactiveUser)

				Expect(err).To(BeNil())
				Expect(event).To(Equal(&user.UserActivated{
					UserID:  inactiveUser.ID.String(),
					Version: 2,
				}))
			})

			Specify("the activated user is persisted in the database", func() {
				inactiveUser, err := user.GetInactive(*db, UserID, nil)
				Expect(err).To(BeNil())

				_, err = user.Activate(*db, *inactiveUser)

				Expect(err).To(BeNil())

				u := user.User{}
				err = db.Model(&u).Where("id = ?", inactiveUser.ID).Take(&u).Error

				Expect(err).To(BeNil())
				Expect(u.ID).To(Equal(inactiveUser.ID))
				Expect(u.EmailAddress).To(Equal("user@example.com"))
				Expect(u.IsActive).To(Equal(true))
				Expect(u.Version).To(Equal(uint32(2)))
			})
		})

		When("the user's state has been modified during deactivation", func() {
			Specify("a state conflict error is returned", func() {
				inactiveUser, err := user.GetInactive(*db, UserID, nil)
				Expect(err).To(BeNil())

				// Simulate a concurrent action on the entity by increasing its version.
				result := db.Model(&user.User{}).
					Where("id = ?", inactiveUser.ID).
					Updates(user.User{
						Version: inactiveUser.Version+1,
					},
					)

				Expect(err).To(BeNil())
				Expect(result.RowsAffected).To(Equal(int64(1)))

				activeUser, err := user.Activate(*db, *inactiveUser)

				Expect(errors.As(err, &domain_errors.StateConflict{})).To(BeTrue())
				Expect(activeUser).To(BeNil())
			})
		})
	})
})
