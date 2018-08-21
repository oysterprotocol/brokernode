package awsgateway

import (
	"encoding/json"

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

	// MaxConcurrency is the number of lambdas running concurrently
	MaxConcurrency = 1000

	// MaxChunksLen is the number of chunks sent to each lambda
	MaxChunksLen = 3000 // 3 MB

	// private
	hooknodeFnName = "arn:aws:lambda:us-east-2:174232317769:function:lambda-node-dev-hooknode"
	hooknodeRegion = "us-east-2"
)

// HooknodeChunk is the chunk object sent to lambda
type HooknodeChunk struct {
	Address string       `json:"address"`
	Value   int          `json:"value"`
	Message giota.Trytes `json:"message"`
	Tag     giota.Trytes `json:"tag"`
}

// HooknodeReq is the payload sent to lambda
type HooknodeReq struct {
	Provider string           `json:"provider"`
	Chunks   []*HooknodeChunk `json:"chunks"`
}

var (
	sess = session.Must(session.NewSession(&aws.Config{Region: aws.String(hooknodeRegion)}))
)

// InvokeHooknode will invoke lambda to do PoW for the chunks in HooknodeReq
func InvokeHooknode(req *HooknodeReq) error {
	// Serialize params
	payload, err := json.Marshal(*req)
	if err != nil {
		return err
	}

	// Invoke lambda.
	client := lambda.New(sess)
	_, err = client.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(hooknodeFnName),
		Payload:      payload,
	})

	return err

	// fmt.Println("=========RESPONSE START=======")
	// fmt.Println("LAMBDA RETURNED")
	// bodyString := string(res.Payload)
	// fmt.Println(bodyString)
	// fmt.Println("========= RESPONSE END =======")

}
