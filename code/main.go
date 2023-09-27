package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MessageResponse struct {
	Counter int64
}

func main() {
	lambda.Start(handler)
}

func handler(request events.ALBTargetGroupRequest) (
	events.ALBTargetGroupResponse, error) {

	fmt.Println("Start Proccess")

	fmt.Println("Body")
	fmt.Println(request.Body)

	messageResponse := MessageResponse{Counter: 0}
	jsonMessageResponse, err := json.Marshal(messageResponse)
	jsonMessageResponseString := string(jsonMessageResponse)

	if err != nil {
		fmt.Println("Erro ao codificar para JSON:", err)
		response := events.ALBTargetGroupResponse{
			StatusCode: 500,
		}
		return response, err
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	response := events.ALBTargetGroupResponse{
		StatusCode:      200,
		Body:            jsonMessageResponseString,
		IsBase64Encoded: false,
		Headers:         headers,
	}

	return response, nil
}
