package transactions

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/farazoman/fo-go-challenge/pkg/files"
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

// summarizes a transaction csv
func (reporter *Reporter) Summarize(sourcePath string) Report {
	transactionsFile := reporter.Loader.Load(sourcePath)
	transactions := extract(transactionsFile)

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
		// TODO verify the int value for month is 1-12s
		if err != nil {
			log.Fatalf("Transaction row value for date has a non-number value for the month position: %s", t.Date)

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
	}
}

func extract(transactionsCsv string) []Transaction {
	var transactions []Transaction
	// TODO make this extensible
	// assume the column order of the CSV is static (always same order)
	for i, row := range strings.Split(transactionsCsv, "\n") {
		if i == 0 || strings.TrimSpace(row) == "" {
			continue
		}

		splitRow := strings.Split(row, ",")

		id, err := strconv.Atoi(strings.TrimSpace(splitRow[0]))
		if err != nil {
			log.Fatalf("%s ID value in file is not an int. With error %s", splitRow[0], err)
		}

		amount, err := strconv.ParseFloat(strings.TrimSpace(splitRow[2]), 32)
		if err != nil {
			log.Fatalf("%s Amount (Transaction) value in file is not a float", splitRow[2])
		}

		transactions = append(transactions,
			Transaction{
				Id:          id,
				Date:        splitRow[1],
				Transaction: float32(amount),
			},
		)
	}

	return transactions
}
