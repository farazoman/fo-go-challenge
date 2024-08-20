# Code Challenge
This challenge processes a transaction file and sends an email to the end user. 

## Run Locally
### Dependencies
1. golang version 1.23 or greater | https://go.dev/doc/install
2. install ginkgo, and ensure the executables are accessible via PATH variable | https://onsi.github.io/ginkgo/#installing-ginkgo
3. run go mod install from root of package
4. ensure AWS credentials are active at ~/.aws

### Overview
This project has no long running service, only serverless instances. To run locally, can only be done via tests.

There are unit and integration tests, to run integration tests, first `tofu apply` must be ran. And locally you need ~/.aws credentials configured and NOT-stale.

### Unit Tests
run:
```
make unit
```

### Integration Tests
These tests still require manual intervention, to update the dynamodb table for the user's email
Then post test to validate the email based on what you have recieved. 
```
make test-integration
```

### To Test On Remote
Build and deploy the code
```
make deploy
```

To test, do the following:
1. Create a new user in Users dynamoDB table. Note the UserId
2. Go to the bucket `fo-notify-transactions`, upload a file with name: `<userId>.1.csv` Copy the content from the file located currently at `pkg/transactions/sample.1.csv`
3. This will trigger the lambda. Verify success by checking the `Transactions` DynamoDB table and your email to make sure that the email arrived.