package actions_v3

func (suite *ActionSuite) Test_GenerateDifferentBucketName() {
	bucketName1 := createUniqueBucketName()
	bucketName2 := createUniqueBucketName()

	suite.NotEqual(bucketName1, bucketName2)
}

func (suite *ActionSuite) Test_CreateAndDeleteBucket() {
	bucket := createUniqueBucketName()
	print(bucket)
	err := createBucket(bucket)

	suite.Nil(err)

	err = deleteBucket(bucket)
	suite.Nil(err)
}

func (suite *ActionSuite) Test_SetAndGetAndDeleteObject() {
	bucket := createUniqueBucketName()
	suite.Nil(createBucket(bucket))

	objectKey := "myKey"
	data := "foo/bar"

	err := setObject(bucket, objectKey, data)
	suite.Nil(err)

	getObjectData, err := getObject(bucket, objectKey, true)

	suite.Nil(err)
	suite.Equal(data, getObjectData)

	err = deleteObject(bucket, objectKey)
	suite.Nil(err)

	suite.Nil(deleteBucket(bucket))
}