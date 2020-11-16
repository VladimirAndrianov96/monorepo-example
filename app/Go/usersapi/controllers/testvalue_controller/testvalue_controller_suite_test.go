package testvalue_controller_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestvalueController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestvalueController Suite")
}
