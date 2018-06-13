package jobs

import (
	"github.com/dgraph-io/badger"
	"github.com/oysterprotocol/brokernode/services"
)

/*BadgerDbGc run garbage collector on current database to do compaction. It will spike the LSTM activity as a result.*/
func BadgerDbGc() {
	db := services.GetBadgerDb()
	if db == nil {
		return
	}

	// It is recommended that this method be called during periods of low activity in your system, or periodically
	for {
		// One call would only result in removal of at max one log file. As an optimization,
		// you could also immediately re-run it whenever it returns nil error.
		// According BadgerDB GoDoc, it is recommended to be 0.5.
		err := db.RunValueLogGC(0.5)
		if err != nil {
			break
		}
	}
}
