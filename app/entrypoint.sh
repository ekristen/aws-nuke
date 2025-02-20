#!/bin/sh

# Define the bucket name and the file key
BUCKET_NAME="heyjobs-aws-nuke-config"
FILE_KEY="config.yml"

cd /app
aws s3 cp s3://$BUCKET_NAME/$FILE_KEY . && chmod 755 /app/*

# Verify the file exists before proceeding
if [ -f "/app/$FILE_KEY" ]; then
  echo "File downloaded successfully."

  # Assume the role and capture the temporary credentials
  CREDENTIALS=$(aws sts assume-role --role-arn arn:aws:iam::867640704278:role/aws_nuke_role --role-session-name aws-nuke-session)

  # Extract the credentials
  export AWS_ACCESS_KEY_ID=$(echo $CREDENTIALS | jq -r '.Credentials.AccessKeyId')
  export AWS_SECRET_ACCESS_KEY=$(echo $CREDENTIALS | jq -r '.Credentials.SecretAccessKey')
  export AWS_SESSION_TOKEN=$(echo $CREDENTIALS | jq -r '.Credentials.SessionToken')

  # Run the aws-nuke command
  aws-nuke nuke -c /app/config.yml --force
  exit 0
else
  echo "File download failed. Exiting."
  exit 1
fi
