package transactions_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransactions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transactions Suite")
}
