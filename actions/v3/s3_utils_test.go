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
