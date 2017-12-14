package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDatadogReporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DatadogReporter Suite")
}
