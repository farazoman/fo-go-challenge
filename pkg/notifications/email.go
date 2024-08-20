package notifications

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Template string

const (
	TransactionSummary Template = "d-b2de268d227d4831a8f95d3b99632689"
	fromEmail          string   = "faraz@farazoman.com"
	fromName           string   = "Faraz Oman"
	sendGridSecretName string   = "SENDGRID_KEY"
)

// Transaction Summary Param Keys
const (
	TotalBalanceKey        = "totalBalance"
	AverageDebitKey        = "averageDebit"
	AverageCreditKey       = "averageCredit"
	TransactionsByMonthKey = "transactionsByMonth"
	MonthKey               = "month"
	AmountKey              = "amount"
)

type Email interface {
	SendEmail(t Template, params EmailParams, toEmail, toName string) error
}

type EmailParams interface {
	ToEmailParams() map[string]interface{}
}

type SendGridNotifier struct {
	awsRegion string
}

func (notifier *SendGridNotifier) SendEmail(t Template, params EmailParams, toEmail, toName string) error {
	apiKey, err := notifier.getSendGridKey()
	if err != nil {
		log.Fatalf("Error getting send grid key %s", err)
	}

	from := mail.NewEmail(fromName, fromEmail)
	to := mail.NewEmail(toName, toEmail)

	msg := mail.NewV3Mail()
	msg.SetTemplateID(string(t))
	msg.SetFrom(from)

	p := mail.NewPersonalization()
	p.AddTos(to)
	for key, value := range params.ToEmailParams() {
		p.SetDynamicTemplateData(key, value)
	}
	msg.AddPersonalizations(p)

	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(msg)
	if err != nil {
		return fmt.Errorf("send failed, see error: %s", err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return nil
}

func (notifier *SendGridNotifier) getSendGridKey() (string, error) {
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(notifier.awsRegion))
	if err != nil {
		return "", err
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(sendGridSecretName),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", err
	}

	return *result.SecretString, nil
}

var _ Email = &SendGridNotifier{}
