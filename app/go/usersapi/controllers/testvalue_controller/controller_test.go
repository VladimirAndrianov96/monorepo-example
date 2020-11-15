package testvalue_controller_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go-ddd-cqrs-example/usersapi/cmd/config"
	"go-ddd-cqrs-example/usersapi/controllers/testvalue_controller"
	"go-ddd-cqrs-example/usersapi/routes"
	"go-ddd-cqrs-example/usersapi/server"
	"go-ddd-cqrs-example/usersapi/server/mocks"
	"io/ioutil"
	"net/http"

	"os"
	"path"
	"runtime"
)

var _ = Describe("Login controller", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *mocks.MockHTTPClient
	)

	mockCtrl = gomock.NewController(GinkgoT())
	mockClient = mocks.NewMockHTTPClient(mockCtrl)

	// Hotfix, fix inconsistent current directory to get configuration file.
	// TODO find better way to handle this.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../..")
	err := os.Chdir(dir)
	Expect(err).To(BeNil())

	// Set up database connection using configuration details.
	cfg := config.Config{}
	viper.AddConfigPath(dir+"/usersapi/cmd/config")
	viper.SetConfigName("configuration")
	viper.ReadInConfig()
	viper.Unmarshal(&cfg)

	srv := server.Server{}
	srv.SecretKey = cfg.SecretKey
	srv.Router = mux.NewRouter()
	srv.Port = cfg.APIAddress
	routes.InitializeRoutes(&srv)

	srv.HTTPClient = mockClient

	Describe("Requesting external API service", func() {
		When("Successful response is returned", func() {
			BeforeEach(func() {
				http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

				testValueResponse := testvalue_controller.TestResponse{"Hello world!"}
				requestBytes, _ := json.Marshal(testValueResponse)
				requestReader := bytes.NewReader(requestBytes)

				httpResponse := http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(requestReader),
				}
				mockClient.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Return(&httpResponse, nil)

			})
			It("returns an OK status and successful response message", func() {
				req, err := http.NewRequest("GET", "https://" + cfg.APIAddress + "/api/get/testvalue", nil)
				Expect(err).To(BeNil())
				Expect(req).ToNot(BeNil())

				client := http.Client{}
				res, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())

				body, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(body).ToNot(BeNil())

				expectedResponse := testvalue_controller.StatusResponse{"Internal connection is fine."}
				reqBodyBytes := new(bytes.Buffer)

				json.NewEncoder(reqBodyBytes).Encode(expectedResponse)

				Expect(body).To(Equal(reqBodyBytes.Bytes()))
			})
		})

		When("Error is returned", func() {
			BeforeEach(func() {
				http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

				httpResponse := http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       nil,
				}
				mockClient.EXPECT().Do(gomock.AssignableToTypeOf(&http.Request{})).Return(&httpResponse, nil)

			})
			It("returns error response", func() {
				req, err := http.NewRequest("GET", "https://" + cfg.APIAddress + "/api/get/testvalue", nil)
				Expect(err).To(BeNil())
				Expect(req).ToNot(BeNil())

				client := http.Client{}
				res, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())

				Expect(res.Status).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})
