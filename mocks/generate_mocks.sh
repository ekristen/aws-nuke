#!/bin/sh

SERVICE=$1
INTERFACE=$2

go run github.com/golang/mock/mockgen -source $(go list -m -mod=mod -f "{{.Dir}}" "github.com/aws/aws-sdk-go")/service/$SERVICE/$INTERFACE/interface.go -destination ../mocks/mock_$INTERFACE/mock.go
