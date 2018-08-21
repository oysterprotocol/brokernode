package awsgateway

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/iotaledger/giota"
)

const (
	// Public
	// https://docs.aws.amazon.com/lambda/latest/dg/limits.html
	// 6MB payload, 300 sec execution time, 1000 concurrent exectutions.
	// Limit to 1000 POSTs and 20 chunks per request.
	MaxConcurrency = 1000
	MaxChunksLen   = 5000 // 3 MB

	// private
	hooknodeFnName = "arn:aws:lambda:us-east-2:174232317769:function:lambda-node-dev-hooknode"
	hooknodeRegion = "us-east-2"
)

type HooknodeChunk struct {
	Address string       `json:"address"`
	Value   int          `json:"value"`
	Message giota.Trytes `json:"message"`
	Tag     giota.Trytes `json:"tag"`
}

type HooknodeReq struct {
	Provider string           `json:"provider"`
	Chunks   []*HooknodeChunk `json:"chunks"`
}

var (
	sess = session.Must(session.NewSession(&aws.Config{Region: aws.String(hooknodeRegion)}))
)

func InvokeHooknode(req *HooknodeReq) error {
	// Serialize params
	payload, err := json.Marshal(*req)
	if err != nil {
		return err
	}

	// Invoke lambda.
	client := lambda.New(sess)
	res, err := client.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(hooknodeFnName),
		Payload:      payload,
	})
	if err != nil {
		return err
	}

	fmt.Println("=========RESPONSE START=======")
	fmt.Println("LAMBDA RETURNED")
	bodyString := string(res.Payload)
	fmt.Println(bodyString)
	fmt.Println("========= RESPONSE END =======")

	return nil
}
