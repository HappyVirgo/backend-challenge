package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type SurveyResponse struct {
	Pk      string `json:"pk"`
	Sk      string `json:"sk"`
	Results []int  `json:"results"`
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Got request: %s\n", request.Body)

	defaultHeaders := make(map[string]string)
	defaultHeaders["Access-Control-Allow-Origin"] = "*"

	data := fetchData()

	body, err := json.Marshal(&data)
	if err != nil {
		fmt.Printf("Failed to marshall json: %v\n", err)

		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    defaultHeaders,
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    defaultHeaders,
		Body:       string(body),
	}, nil
}

func fetchData() map[string][]float32 {
	result := map[string][]float32{}
	var mutex sync.Mutex // Add mutex for protecting map access

	departments := []string{
		"SAAS development",
		"Bespoke development",
		"Creative",
		"HR",
		"QA",
		"Project Management",
	}

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(departments))

	for _, dept := range departments {
		go fetchDepartmentData(&waitGroup, &result, &mutex, dept) // Pass mutex to goroutine
	}

	waitGroup.Wait()

	return result
}
func fetchDepartmentData(waitGroup *sync.WaitGroup, target *map[string][]float32, mutex *sync.Mutex, dept string) {
	defer waitGroup.Done()
	result := []float32{0, 0, 0, 0}

	scanInput := dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("TableName")),
		ScanFilter: map[string]*dynamodb.Condition{
			"sk": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(fmt.Sprintf("DEPARTMENT#%s", dept)),
					},
				},
			},
		},
	}

	svc := dynamodb.New(session.New())
	scanResult, err := svc.Scan(&scanInput)
	if err != nil {
		fmt.Printf("Failed to fetch items: %v\n", err)
	} else {
		var responses []SurveyResponse
		err = dynamodbattribute.UnmarshalListOfMaps(scanResult.Items, &responses)

		for _, response := range responses {
			for i := 0; i < 4; i++ {
				result[i] += float32(response.Results[i])
			}
		}

		if len(responses) > 0 {
			responseCount := float32(len(responses))
			result[0] /= responseCount
			result[1] /= responseCount
			result[2] /= responseCount
			result[3] /= responseCount
		}
	}

	// Protect map write with mutex
	mutex.Lock()
	(*target)[dept] = result
	mutex.Unlock()
}
