package transactions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/farazoman/fo-go-challenge/pkg/files"
	"github.com/farazoman/fo-go-challenge/pkg/notifications"
)

type Transaction struct {
	Id          int
	Date        string
	Transaction float32
}

type Reporter struct {
	Loader files.Loader
}

type Report struct {
	TotalBalance         float32
	TransactionsPerMonth map[string]int
	AverageDebit         float32
	AverageCredit        float32
}

func (r Report) String() string {
	return fmt.Sprintf(`
Total Balance: %f
Transactions Per Month: %v
Average Debit: %f
Average Credit: %f`,
		r.TotalBalance, r.TransactionsPerMonth, r.AverageDebit, r.AverageCredit)
}

// summarizes a transaction csv
func (reporter *Reporter) Summarize(sourcePath string) (Report, []Transaction, error) {
	transactionsFile, err := reporter.Loader.Load(sourcePath)
	if err != nil {
		return Report{}, []Transaction{}, err
	}
	transactions, err := extract(transactionsFile)
	if err != nil {
		return Report{}, transactions, err
	}

	var totalCredit float32
	var creditCount int
	var totalDebit float32
	var debitCount int
	perMonth := make(map[string]int, 12)
	for _, t := range transactions {
		if t.Transaction < 0 {
			totalCredit += t.Transaction
			creditCount++
		} else {
			totalDebit += t.Transaction
			debitCount++
		}

		monthDay := strings.Split(t.Date, "/")
		month, err := strconv.Atoi(monthDay[0])

		if err != nil {
			return Report{}, []Transaction{},
				fmt.Errorf("Transaction row value for date has a non-number value for the month position: %s", t.Date)
		}
		if month < 1 || month > 12 {
			return Report{}, []Transaction{},
				fmt.Errorf("number for month needs to be between 1 and 12. Value was: %d", month)
		}

		date := time.Date(2024, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		monthString := date.Month().String()

		perMonth[monthString] = perMonth[monthString] + 1
	}

	return Report{
		TotalBalance:         totalCredit + totalDebit,
		AverageCredit:        totalCredit / float32(creditCount),
		AverageDebit:         totalDebit / float32(debitCount),
		TransactionsPerMonth: perMonth,
	}, transactions, nil
}

func validateHeader(row string) error {
	headers := strings.Split(row, ",")
	if strings.ToLower(strings.TrimSpace(headers[0])) != "id" ||
		strings.ToLower(strings.TrimSpace(headers[1])) != "date" ||
		strings.ToLower(strings.TrimSpace(headers[2])) != "transaction" {
		return fmt.Errorf("headers are not valid, should be ID then Date, then Transaction. The header row was %s", row)
	}
	return nil
}

func extract(transactionsCsv string) ([]Transaction, error) {
	var transactions []Transaction
	// assume the column order of the CSV is static (always same order)
	for i, row := range strings.Split(transactionsCsv, "\n") {
		if i == 0 {
			err := validateHeader(row)
			if err != nil {
				return []Transaction{}, err
			}
			continue
		}
		if strings.TrimSpace(row) == "" {
			continue
		}

		splitRow := strings.Split(row, ",")

		id, err := strconv.Atoi(strings.TrimSpace(splitRow[0]))
		if err != nil {
			return []Transaction{}, fmt.Errorf("%s ID value in file is not an int. With error %s", splitRow[0], err)
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(splitRow[2]), 32)
		if err != nil {
			return []Transaction{}, fmt.Errorf("%s Amount (Transaction) value in file is not a float", splitRow[2])
		}

		transactions = append(transactions,
			Transaction{
				Id:          id,
				Date:        splitRow[1],
				Transaction: float32(amount),
			},
		)
	}

	return transactions, nil
}

func (r *Report) ToEmailParams() map[string]interface{} {
	params := make(map[string]interface{})
	params[notifications.TotalBalanceKey] = strconv.FormatFloat(float64(r.TotalBalance), 'f', 2, 32)
	params[notifications.AverageDebitKey] = strconv.FormatFloat(float64(r.AverageDebit), 'f', 2, 32)
	params[notifications.AverageCreditKey] = strconv.FormatFloat(float64(r.AverageCredit), 'f', 2, 32)

	allByMonth := make([]map[string]string, len(r.TransactionsPerMonth))

	var i int
	for month, amount := range r.TransactionsPerMonth {
		byMonth := make(map[string]string)
		byMonth[notifications.AmountKey] = strconv.Itoa(amount)
		byMonth[notifications.MonthKey] = month
		allByMonth[i] = byMonth
		i++
	}

	params[notifications.TransactionsByMonthKey] = allByMonth
	return params
}
