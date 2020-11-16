package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nsqio/go-nsq"
	"github.com/spf13/viper"
	"go-ddd-cqrs-example/domain/models/user"
	"go-ddd-cqrs-example/usersapi/cmd/config"
	"go-ddd-cqrs-example/usersapi/routes"
	"go-ddd-cqrs-example/usersapi/server"
	"go-ddd-cqrs-example/usersapi/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
)

var apiServer = server.Server{}
var cfg config.Config

// initLogger initializes the zap logger with reasonable
// defaults and replaces the global logger.
func initLogger() error {
	// Initialize the logs encoder.
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeDuration = zapcore.StringDurationEncoder

	// Initialize the logger.
	logger, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    encoder,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		return err
	}

	// Then replace the globals.
	zap.ReplaceGlobals(logger)

	return nil
}

func loadConfiguration() error {
	// Load up configuration.
	viper.AddConfigPath("./usersapi/cmd/config")
	viper.SetConfigName("configuration")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}

	return nil
}

// initialize the database connection and the HTTP router.
func initializeAPI(server *server.Server, driver, username, password, port, host, database string) error {
	var err error

	server.DB, err = utils.GetDB(driver, username, password, port, host, database)
	if err != nil {
		return err
	}

	// Database migration
	server.DB.AutoMigrate(
		&user.User{},
	)

	server.Router = mux.NewRouter()
	routes.InitializeRoutes(server)
	server.HTTPClient = &http.Client{}

	return nil
}

func main() {
	// Disable cert verification to use self-signed certificates for internal service needs.
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Global logging synchronizer.
	// This ensures the logged data is flushed out of the buffer before program exits.
	defer zap.S().Sync()

	err := initLogger()
	if err != nil {
		zap.S().Fatal(err)
	}

	err = loadConfiguration()
	if err != nil {
		zap.S().Fatal(err)
	}

	nsqConfig := nsq.NewConfig()

	//Creating the Producer using NSQD Address
	producer, err := nsq.NewProducer("nsqd:4150", nsqConfig)
	if err != nil {
		log.Fatal(err)
	}

	srv := server.Server{}
	srv.Port = cfg.APIAddress
	srv.SecretKey = cfg.SecretKey
	srv.TestAPIAddress = cfg.TestAPIAddress
	srv.EventEmitter = *producer

	err = initializeAPI(
		&srv,
		cfg.DBDriver,
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBPort,
		cfg.DBHost,
		cfg.DBName,
	)
	if err != nil {
		zap.S().Fatal(err)
	}

	err = run(&srv, srv.Port)
	if err != nil {
		zap.S().Fatal(err)
	}
}

func run(server *server.Server, addr string) error {
	defer server.DB.Close()

	fmt.Println("Listening to " + addr)
	err := http.ListenAndServeTLS(addr,
		"./usersapi/golangbackend.crt",
		"./usersapi/golangbackend.key",
		handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Accept", "Accept-Language"}),
			handlers.AllowedMethods([]string{"GET", "POST"}),
			handlers.AllowedOrigins([]string{"*"}),
		)(server.Router))
	if err != nil {
		return err
	}

	return nil
}
