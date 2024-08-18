package transactions_test

import (
	"github.com/farazoman/fo-go-challenge/pkg/files"
	"github.com/farazoman/fo-go-challenge/pkg/transactions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Report Summary", func() {
	var (
		reporter     transactions.Reporter
		actualReport transactions.Report
	)
	BeforeEach(func() {
		reporter = transactions.Reporter{Loader: files.MockLoader{}}
		actualReport = reporter.Summarize("anyPath")
	})

	When("A file with 2 entries in two months is loaded", func() {
		It("A report is generated with the two months' info and other report info correctly returned", func() {
			expectedReport := Fields{
				"TotalBalance":         BeNumerically("~", 13.47, 0.01),
				"AverageDebit":         BeNumerically("~", 4.86, 0.01),
				"AverageCredit":        BeNumerically("~", -1.1, 0.01),
				"TransactionsPerMonth": Equal(map[string]int{"September": 2, "August": 2}),
			}
			Expect(actualReport).Should(MatchAllFields(expectedReport))
		})
	})
})
