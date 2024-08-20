deploy:
	GOOS=linux go build -tags lambda.norpc -o out/bootstrap lambda/main.go
	tofu apply --auto-approve

unit:
	ginkgo run -r --skip-package integration

test-integration: 
	echo "Ensure AWS has been configured, read integration test spec for more setup details"
	sleep 5
	ginkgo run -r integration