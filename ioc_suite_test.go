package ioc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIoc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ioc Suite")
}
