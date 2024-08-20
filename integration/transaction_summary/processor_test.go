package transaction_summary_test

import (
	"github.com/farazoman/fo-go-challenge/pkg/files"
	"github.com/farazoman/fo-go-challenge/pkg/transactions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test requires manual intervention and validation
// 1. in dynamoDB go to "Users" table and create a user with some unique key. Use your email
// 2. using the unique key, rename the file in the transactions package from `sample.1.csv` to `<uniqueKey>.1.csv`
// 3. ensure that your cedentials in ~/.aws is up to date
// 4. run tests with ginkgo -r from within this directory
// 5. validate the results in your email
var _ = Describe("Report Summary", func() {
	var (
		reporter    transactions.Reporter
		actualError error
	)
	BeforeEach(func() {
		reporter = transactions.Reporter{Loader: files.S3SingleBucketLoader{Bucket: "fo-notify-transactions"}}
		actualError = transactions.Process("inbox/sample.1.csv", reporter)
	})

	When("A file with 2 entries in two months is loaded", func() {
		It("Does not return an error", func() {
			Expect(actualError).Should(BeNil())
		})
	})
})
