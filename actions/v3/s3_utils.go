package actions_v3

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/oysterprotocol/brokernode/utils"
)

var Svc *s3.S3
var bucketPrefix string
var counter uint64

func init() {
	sess := session.Must(session.NewSession())
	Svc = s3.New(sess)
	if v := os.Getenv("DISPLAY_NAME"); v != "" {
		bucketPrefix = v
	} else {
		bucketPrefix = "unknown"
	}
	bucketPrefix = strings.Replace(bucketPrefix, "_", "-", -1)
}

/* Create unique bucket name. */
func createUniqueBucketName() string {
	atomic.AddUint64(&counter, 1)
	return strings.ToLower(fmt.Sprintf("%v-%v-%v", bucketPrefix, time.Now().Format("2006-01-02t15.04.05z07.00"), counter))
}

/* Create a private bucket with bucketName. */
func createBucket(bucketName string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := Svc.CreateBucket(input)
	oyster_utils.LogIfError(err, nil)
	return err
}

/* Delete bucket as bucketName. */
func deleteBucket(bucketName string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := Svc.DeleteBucket(input)
	oyster_utils.LogIfError(err, nil)
	return err
}
