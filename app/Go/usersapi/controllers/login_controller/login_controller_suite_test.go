package login_controller_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoginController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LoginController Suite")
}
