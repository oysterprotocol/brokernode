package jobs_test

/*
import (
	"github.com/oysterprotocol/brokernode/jobs"
	"github.com/oysterprotocol/brokernode/utils"
	"time"
)

func (suite *JobsSuite) Test_BadgerDbGc() {

	keyToDelete := oyster_utils.RandSeq(6, []rune("abcdefghijklmnopqrstuvwxyz"))
	valueToDelete := "this_should_get_deleted"

	oyster_utils.BatchSet(&oyster_utils.KVPairs{keyToDelete: valueToDelete},
		1*time.Second)

	keyToKeep := oyster_utils.RandSeq(6, []rune("abcdefghijklmnopqrstuvwxyz"))
	valueToKeep := "this_should_NOT_get_deleted"

	oyster_utils.BatchSet(&oyster_utils.KVPairs{keyToKeep: valueToKeep},
		10*time.Minute)

	keyValuePair, err := oyster_utils.BatchGet(&oyster_utils.KVKeys{keyToKeep, keyToDelete})
	suite.Nil(err)

	suite.Equal(valueToKeep, (*(keyValuePair))[keyToKeep])
	suite.Equal(valueToDelete, (*(keyValuePair))[keyToDelete])

	time.Sleep(3 * time.Second)

	jobs.BadgerDbGc()

	time.Sleep(500 * time.Millisecond)

	keyValuePair, err = oyster_utils.BatchGet(&oyster_utils.KVKeys{
		keyToKeep,
		keyToDelete,
	})
	suite.Nil(err)

	//Should have been garbage collected
	suite.NotEqual(valueToDelete, (*(keyValuePair))[keyToDelete])

	//Should NOT have been garbage collected
	suite.Equal(valueToKeep, (*(keyValuePair))[keyToKeep])
}
*/
