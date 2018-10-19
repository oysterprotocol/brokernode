package actions_v3

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/orcaman/concurrent-map"
	"github.com/oysterprotocol/brokernode/utils"
)

var svc *s3.S3
var bucketPrefix string
var counter uint64

var cachedData cmap.ConcurrentMap

func init() {
	sess := session.Must(session.NewSession())
	svc = s3.New(sess)
	if v := os.Getenv("DISPLAY_NAME"); v != "" {
		bucketPrefix = v
	} else {
		bucketPrefix = "unknown"
	}
	bucketPrefix = strings.Replace(bucketPrefix, "_", "-", -1)

	cachedData = cmap.New()
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
	_, err := svc.CreateBucket(input)
	oyster_utils.LogIfError(err, nil)
	return err
}

/* Delete bucket as bucketName. Must make sure no object inside the bucket*/
func deleteBucket(bucketName string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := svc.DeleteBucket(input)
	oyster_utils.LogIfError(err, nil)
	return err
}

func getObject(bucketName string, objectKey string, cached bool) (string, error) {
	if cached {
		if value, ok := cachedData.Get(getKey(bucketName, objectKey)); ok {
			return value.(string), nil
		}
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	output, err := svc.GetObject(input)
	oyster_utils.LogIfError(err, nil)

	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)
	return buf.String(), nil
}

func setObject(bucketName string, objectKey string, data string) error {
	input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(strings.NewReader(data)),
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	_, err := svc.PutObject(input)
	oyster_utils.LogIfError(err, nil)

	if err != nil {
		cachedData.Set(getKey(bucketName, objectKey), data)
	}
	return err
}

func deleteObject(bucketName string, objectKey string) error {
	cachedData.Remove(getKey(bucketName, objectKey))

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}

	_, err := svc.DeleteObject(input)
	oyster_utils.LogIfError(err, nil)
	return err
}

func getKey(bucketName string, objectKey string) string {
	return fmt.Sprintf("%v:%v", bucketName, objectKey)
}
