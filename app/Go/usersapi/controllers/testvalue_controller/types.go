package testvalue_controller

type TestResponse struct {
	Value  string `json:"value"`
}

type StatusResponse struct{
	Message string `json:"response"`
}