package transactions

import (
	"fmt"
	"strings"

	"github.com/farazoman/fo-go-challenge/pkg/notifications"
	"github.com/farazoman/fo-go-challenge/pkg/storage"
)

func getUserId(filePath string) (string, error) {
	// assume all files have prefix /inbox
	fileComponents := strings.Split(strings.Split(filePath, "/")[1], ".")
	if len(fileComponents) != 3 {
		return "", fmt.Errorf("file path '%s' does not follow the given format of '<userId>.<number>.csv'", filePath)
	}
	return fileComponents[0], nil
}

func getUser(filePath string) (storage.User, error) {
	userId, err := getUserId(filePath)
	if err != nil {
		return storage.User{}, err
	}

	return storage.GetItem[storage.User](
		storage.GenKeyS(storage.UserIdKey, userId),
		"Users",
	)
}

func Process(filePath string, r Reporter) error {
	fmt.Printf("Summarizing the transaction for file %s\n", filePath)
	report, transactions, err := r.Summarize(filePath)
	if err != nil {
		return err
	}
	fmt.Printf("DEBUG: %s\n", report)

	user, err := getUser(filePath)
	if err != nil {
		fmt.Println("Failed getting user ID")
		return err
	}

	// Note: DB insertion and sending email should be decoupled to prevent duplicates
	// We could implement dedupe logic so that an email only gets sent once, or;
	// the db transaction record uses the same partition key as to not create new records
	fmt.Printf("Saving transactions to db. Number of transactions %d\n", len(transactions))
	err = saveTransactions(transactions, &user)
	if err != nil {
		return err
	}

	fmt.Printf("Sending summary to email %s\n", user.Email)
	n := notifications.SendGridNotifier{}
	return n.SendEmail(
		notifications.TransactionSummary,
		&report,
		user.Email,
		fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	)
}

func saveTransactions(transactions []Transaction, user *storage.User) error {
	for _, t := range transactions {
		err := storage.InsertTransaction(storage.Transaction{
			UserId:        user.UserId,
			TransactionId: t.Id,
			Date:          t.Date,
			Transaction:   t.Transaction,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
