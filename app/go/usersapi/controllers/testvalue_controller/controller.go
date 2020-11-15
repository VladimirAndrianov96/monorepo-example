package testvalue_controller

import (
	"encoding/json"
	"go-ddd-cqrs-example/usersapi/responses"
	"go-ddd-cqrs-example/usersapi/server"
	"io/ioutil"
	"net/http"
)

// GetTestValue from API to ensure the internal communication between services works fine.
func GetTestValue(server *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequest("GET", "http://" + server.TestAPIAddress + "/api/get/testvalue", nil)
		if err != nil{
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		res, err := server.HTTPClient.Do(req)
		if err != nil{
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil{
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		testValueResponse := TestResponse{}
		err = json.Unmarshal(body, &testValueResponse)
		if err != nil{
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		if testValueResponse.Value != "Hello world!" {
			responses.ERROR(w, http.StatusInternalServerError, nil)
			return
		}

		responses.JSON(w, http.StatusOK, StatusResponse{"Internal connection is fine."})
	}
}
