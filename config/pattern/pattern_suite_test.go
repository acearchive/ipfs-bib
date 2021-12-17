package pattern_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestPattern(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pattern Suite")
}
