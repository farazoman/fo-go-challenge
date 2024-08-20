terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.63"
    }
  }
}

variable "aws_region" {
  default = "us-west-2"
  type = string
}

provider "aws" {
  region     = var.aws_region
  access_key = ""
  secret_key = ""
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "iam_for_lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

variable "zip_path" {
   default = "out/zip/lambda.zip"
   type = string
}

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "out/bootstrap"
  output_path = var.zip_path
}

resource "aws_lambda_function" "notify_lambda" {
  filename      = "${path.module}/${var.zip_path}"
  function_name = "transaction_summary_notifier"
  role          = aws_iam_role.iam_for_lambda.arn 
  handler       = "bootstrap"
  
  source_code_hash = data.archive_file.lambda.output_base64sha256

  runtime = "provided.al2023"

  environment {
    variables = {
      NOTIFY_BUCKET_NAME = aws_s3_bucket.notify_s3.id
      TO_EMAIL           = "farazoman@gmail.com"
    }
  }
  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.group,
  ]
}

resource "aws_s3_bucket" "notify_s3" {
  bucket = "fo-notify-transactions"
  force_destroy = true
}

resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.notify_lambda.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.notify_s3.arn
}


resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.notify_s3.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.notify_lambda.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "inbox/"
    filter_suffix       = ".csv"
  }

  depends_on = [aws_lambda_permission.allow_bucket]
}

data "aws_iam_policy_document" "lambda_s3_read" {
  statement {
    effect = "Allow"

    actions = [
      "s3:ListBucket",
      "s3:GetObject",
    ]

    resources = [ "${aws_s3_bucket.notify_s3.arn}/inbox/*", aws_s3_bucket.notify_s3.arn ]
  }
}

resource "aws_iam_policy" "lambda_s3_read" {
  name        = "lambda_s3_read"
  path        = "/"
  description = "IAM policy for reading inbox from notify s3 bucket"
  policy      = data.aws_iam_policy_document.lambda_s3_read.json
}

resource "aws_iam_role_policy_attachment" "lambda_s3_read" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_s3_read.arn
}


// Logging Config
variable "lambda_function_name" {
  default = "lambda_function_name"
}

resource "aws_cloudwatch_log_group" "group" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 1
}

data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

// Access to Secrets
data "aws_iam_policy_document" "lambda_secrets" {
  statement {
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
    ]

    resources = [ data.aws_secretsmanager_secret.sendgrid_secret.arn ]
  }
}

data "aws_secretsmanager_secret" "sendgrid_secret" {
  name = "SENDGRID_KEY"
}

resource "aws_iam_policy" "lambda_secrets" {
  name        = "lambda_secrets"
  path        = "/"
  description = "IAM policy for getting secret value from a lambda"
  policy      = data.aws_iam_policy_document.lambda_secrets.json
}


resource "aws_iam_role_policy_attachment" "lambda_secrets" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_secrets.arn
}

// Dynamo DBs
resource "aws_dynamodb_table" "transaction_table" {
  name           = "Transactions"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "GlobalTransactionId"

  attribute {
    name = "GlobalTransactionId"
    type = "S"
  }
}

resource "aws_dynamodb_table" "user_table" {
  name           = "Users"
  billing_mode   = "PAY_PER_REQUEST"

  hash_key       = "UserId"

  attribute {
    name = "UserId"
    type = "S"
  }
}

data "aws_iam_policy_document" "lambda_dynamo" {
  statement {
    effect = "Allow"

    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
    ]

    resources = [
      aws_dynamodb_table.transaction_table.arn,
      aws_dynamodb_table.user_table.arn
    ]
  }
}

resource "aws_iam_policy" "lambda_dynamo" {
  name        = "lambda_dynamo"
  path        = "/"
  description = "IAM policy for lambda to access dynamo"
  policy      = data.aws_iam_policy_document.lambda_dynamo.json
}

resource "aws_iam_role_policy_attachment" "lambda_dynamo" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_dynamo.arn
}