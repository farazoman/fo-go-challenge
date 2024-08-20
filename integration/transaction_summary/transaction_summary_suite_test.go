package transaction_summary_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransactionSummary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TransactionSummary Suite")
}
