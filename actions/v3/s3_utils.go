package actions_v3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/oysterprotocol/brokernode/utils"
)

var Svc *s3.S3

func init() {
	sess := session.Must(session.NewSession())
	Svc = s3.New(sess)
}

/* Create unique bucket name. */
func createUniqueBucketName() string {
	// TODO: Figure out how to make this unique
	return ""
}

/* Create a private bucket with bucketName. */
func createBucket(bucketName string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	result, err := Svc.CreateBucket(input)
	if err == nil {
		fmt.Println(result)
	}
	oyster_utils.LogIfError(err, nil)
	return err
}

/* Delete bucket as bucketName. */
func deleteBucket(bucketName string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	result, err := Svc.DeleteBucket(input)
	if err == nil {
		fmt.Println(result)
	}
	oyster_utils.LogIfError(err, nil)
	return err
}
