package actions_v3

func (suite *ActionSuite) Test_GenerateDifferentBucketName() {
	bucketName1 := createUniqueBucketName()
	bucketName2 := createUniqueBucketName()

	suite.NotEqual(bucketName1, bucketName2)
}

func (suite *ActionSuite) Test_CreateAndDeleteBucket() {
	if !isS3Enabled() {
		return
	}

	bucket := createUniqueBucketName()
	print(bucket)
	err := createBucket(bucket)

	suite.Nil(err)

	err = deleteBucket(bucket)
	suite.Nil(err)
}

func (suite *ActionSuite) Test_SetAndGetAndDeleteObject() {
	if !isS3Enabled() {
		return
	}

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

func (suite *ActionSuite) Test_ListObject() {
	if !isS3Enabled() {
		return
	}

	bucket := createUniqueBucketName()
	suite.Nil(createBucket(bucket))

	objectKeys := []string{"o1", "o2", "o3", "o4", "o5"}
	for _, k := range objectKeys {
		suite.Nil(setObject(bucket, k, "data"))
	}

	awsPagingSize = 2

	l, err := listObjectKeys(bucket, "o")

	suite.Nil(err)
	suite.Equal(objectKeys, l)

	awsPagingSize = 1000

	// Clean up the bucket
	for _, k := range objectKeys {
		suite.Nil(deleteObject(bucket, k))
	}
	suite.Nil(deleteBucket(bucket))
}

func (suite *ActionSuite) Test_BatchDelete() {
	if !isS3Enabled() {
		return
	}

	awsPagingSize = 2

	bucket := createUniqueBucketName()
	suite.Nil(createBucket(bucket))

	objectKeys := []string{"o/1", "o/2", "o/3", "o/4", "o/5"}
	for _, k := range objectKeys {
		suite.Nil(setObject(bucket, k, "data"))
	}

	l, err := listObjectKeys(bucket, "o/")
	suite.Nil(err)
	suite.Equal(objectKeys, l)

	err = deleteObjectKeys(bucket, "o/")
	suite.Nil(err)

	l, err = listObjectKeys(bucket, "o/")
	suite.Nil(err)
	suite.True(len(l) == 0)

	suite.Nil(deleteBucket(bucket))

	awsPagingSize = 1000
}
