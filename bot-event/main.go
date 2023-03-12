package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type At struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
	AtUserIds []string `json:"atUserIds"`
}

type Text struct {
	Content string `json:"content"`
}

type Link struct {
	MessageUrl string `json:"messageUrl"`
	PicUrl     string `json:"picUrl"`
	Title      string `json:"title"`
	Text       string `json:"text"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type OapiRobotSendRequest struct {
	MsgType  string   `json:"msgtype"`
	Text     Text     `json:"text,omitempty"`
	Link     Link     `json:"link,omitempty"`
	Markdown Markdown `json:"markdown,omitempty"`
	At       At       `json:"at,omitempty"`
}

type OapiRobotSendResponse struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func Handler(ctx context.Context, snsEvent events.SNSEvent) error {
	snsMsg := snsEvent.Records[0].SNS.Message

	req := OapiRobotSendRequest{
		MsgType: "text",
		Text: Text{
			Content: "AWS Message:\n" + snsMsg,
		},
		At: At{
			IsAtAll: false,
		},
	}

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	secretARN := os.Getenv("SECRET_ARN")

	if secretARN == "" {
		return fmt.Errorf("secret ARN is not set in environment variable: SECRET_ARN")
	}

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	svc := secretsmanager.NewFromConfig(cfg)

	output, err := svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})

	if err != nil {
		return fmt.Errorf("failed to get secret value: %w", err)
	}

	secretValue := aws.ToString(output.SecretString)
	if secretValue == "" {
		return fmt.Errorf("secret value is empty")
	}

	httpReq, err := http.NewRequest("POST", secretValue, bytes.NewBuffer(jsonReq))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	var jsonResp OapiRobotSendResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}

	if jsonResp.Errcode != 0 {
		return fmt.Errorf("error sending message: %d %s", jsonResp.Errcode, jsonResp.Errmsg)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
