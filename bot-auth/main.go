package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Signature string `json:"signature"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Encrypt   string `json:"encrypt"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	signature := request.QueryStringParameters["signature"]
	timestamp := request.QueryStringParameters["timestamp"]
	nonce := request.QueryStringParameters["nonce"]
	encrypt := request.Body

	response := Response{
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     nonce,
		Encrypt:   encrypt,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error marshalling response body", StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(responseBody),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
