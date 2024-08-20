package storage

import (
	"context"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const (
	GlobalTransactionIdKey = "GlobalTransactionId"
	UserIdTransactionKey   = "UserId"
	TransactionIdKey       = "TransactionId"
	DateKey                = "Date"
	TransactionKey         = "TransactionKey"

	TransactionsTableName = "Transactions"
)

type Transaction struct {
	GlobalTransactionId string
	UserId              string
	TransactionId       int
	Date                string
	Transaction         float32
}

const (
	UserIdKey    = "UserId"
	FirstNameKey = "FirstName"
	LastNameKey  = "LastName"
	EmailKey     = "Email"
)

type User struct {
	UserId    string
	FirstName string
	LastName  string
	Email     string
}

func InsertTransaction(transaction Transaction) error {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	client := dynamodb.NewFromConfig(cfg)

	_, err := client.PutItem(
		context.TODO(),
		&dynamodb.PutItemInput{
			TableName: aws.String(TransactionsTableName),
			Item: map[string]types.AttributeValue{
				// NOTE: the global transaction key will not save indempodently
				// This means reprocessing a file will cause duplicates in the db
				// To solve, the global id needs to be generated deterministically and be unique
				GlobalTransactionIdKey: &types.AttributeValueMemberS{Value: uuid.NewString()},
				UserIdTransactionKey:   &types.AttributeValueMemberS{Value: transaction.UserId},
				TransactionIdKey:       &types.AttributeValueMemberN{Value: strconv.Itoa(transaction.TransactionId)},
				DateKey:                &types.AttributeValueMemberS{Value: transaction.Date},
				TransactionKey:         &types.AttributeValueMemberN{Value: strconv.FormatFloat(float64(transaction.Transaction), 'f', -1, 32)},
			},
		},
	)
	return err
}

func GenKeyS(keyName, keyValue string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		keyName: &types.AttributeValueMemberS{Value: keyValue},
	}
}

func GetItem[T interface{}](key map[string]types.AttributeValue, tableName string) (T, error) {
	var item T
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	client := dynamodb.NewFromConfig(cfg)

	response, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(tableName),
	})
	if err != nil {
		log.Printf("Couldn't get info about %v. Here's why: %v\n", key, err)
		return item, err
	}

	err = attributevalue.UnmarshalMap(response.Item, &item)
	if err != nil {
		log.Printf("Couldn't unmarshal response. Here's why: %v\n", err)
	}

	return item, err
}
