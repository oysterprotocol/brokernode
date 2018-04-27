package jobs

import (
	"github.com/gobuffalo/buffalo/worker"
	"github.com/oysterprotocol/brokernode/services"
	"time"
)

var BundleSize = 30

var OysterWorker = worker.NewSimple()

var IotaWrapper = services.IotaWrapper
var EthWrapper = services.EthWrapper

func init() {
	registerHandlers(OysterWorker)

	doWork(OysterWorker)
}

func registerHandlers(oysterWorker *worker.Simple) {
	oysterWorker.Register("flushOldWebnodesHandler", flushOldWebnodesHandler)
	oysterWorker.Register("processUnassignedChunksHandler", processUnassignedChunksHandler)
	oysterWorker.Register("purgeCompletedSessionsHandler", purgeCompletedSessionsHandler)
	oysterWorker.Register("verifyDataMapsHandler", verifyDataMapsHandler)
	oysterWorker.Register("updateTimedOutDataMapsHandler", updateTimedOutDataMapsHandler)
	oysterWorker.Register("processPaidSessionsHandler", processPaidSessionsHandler)
	oysterWorker.Register("claimUnusedPRLsHandler", claimUnusedPRLsHandler)
}

func doWork(oysterWorker *worker.Simple) {
	flushOldWebnodesJob := worker.Job{
		Queue:   "default",
		Handler: "flushOldWebnodesHandler",
		Args: worker.Args{
			"duration": 5 * time.Minute,
		},
	}

	processUnassignedChunksJob := worker.Job{
		Queue:   "default",
		Handler: "processUnassignedChunksHandler",
		Args: worker.Args{
			"duration": 20 * time.Second,
		},
	}

	purgeCompletedSessionsJob := worker.Job{
		Queue:   "default",
		Handler: "purgeCompletedSessionsHandler",
		Args: worker.Args{
			"duration": 60 * time.Second,
		},
	}

	verifyDataMapsJob := worker.Job{
		Queue:   "default",
		Handler: "verifyDataMapsHandler",
		Args: worker.Args{
			"duration": 30 * time.Second,
		},
	}

	updateTimedOutDataMapsJob := worker.Job{
		Queue:   "default",
		Handler: "updateTimedOutDataMapsHandler",
		Args: worker.Args{
			"duration": 60 * time.Second,
		},
	}

	processPaidSessionsJob := worker.Job{
		Queue:   "default",
		Handler: "processPaidSessionsHandler",
		Args: worker.Args{
			"duration": 30 * time.Second,
		},
	}

	claimUnusedPRLsJob := worker.Job{
		Queue:   "default",
		Handler: "claimUnusedPRLsHandler",
		Args: worker.Args{
			"duration": 10 * time.Minute,
		},
	}

	oysterWorker.PerformIn(flushOldWebnodesJob, flushOldWebnodesJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(processUnassignedChunksJob, processUnassignedChunksJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(purgeCompletedSessionsJob, purgeCompletedSessionsJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(verifyDataMapsJob, verifyDataMapsJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(updateTimedOutDataMapsJob, updateTimedOutDataMapsJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(processPaidSessionsJob, processPaidSessionsJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(claimUnusedPRLsJob, claimUnusedPRLsJob.Args["duration"].(time.Duration))
}

var flushOldWebnodesHandler = func(args worker.Args) error {
	thresholdTime := time.Now().Add(-20 * time.Minute) // webnodes older than 20 minutes get deleted
	FlushOldWebNodes(thresholdTime)

	flushOldWebnodesJob := worker.Job{
		Queue:   "default",
		Handler: "flushOldWebnodesHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(flushOldWebnodesJob, flushOldWebnodesJob.Args["duration"].(time.Duration))

	return nil
}

var processUnassignedChunksHandler = func(args worker.Args) error {
	ProcessUnassignedChunks(IotaWrapper)

	processUnassignedChunksJob := worker.Job{
		Queue:   "default",
		Handler: "processUnassignedChunksHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(processUnassignedChunksJob, processUnassignedChunksJob.Args["duration"].(time.Duration))

	return nil
}

var purgeCompletedSessionsHandler = func(args worker.Args) error {
	PurgeCompletedSessions()

	purgeCompletedSessionsJob := worker.Job{
		Queue:   "default",
		Handler: "purgeCompletedSessionsHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(purgeCompletedSessionsJob, purgeCompletedSessionsJob.Args["duration"].(time.Duration))

	return nil
}

var verifyDataMapsHandler = func(args worker.Args) error {
	VerifyDataMaps(IotaWrapper)

	verifyDataMapsJob := worker.Job{
		Queue:   "default",
		Handler: "verifyDataMapsHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(verifyDataMapsJob, verifyDataMapsJob.Args["duration"].(time.Duration))

	return nil
}

var updateTimedOutDataMapsHandler = func(args worker.Args) error {
	UpdateTimeOutDataMaps(time.Now().Add(-1 * time.Minute))

	updateTimedOutDataMapsJob := worker.Job{
		Queue:   "default",
		Handler: "updateTimedOutDataMapsHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(updateTimedOutDataMapsJob, updateTimedOutDataMapsJob.Args["duration"].(time.Duration))

	return nil
}

var processPaidSessionsHandler = func(args worker.Args) error {
	ProcessPaidSessions()

	processPaidSessionsJob := worker.Job{
		Queue:   "default",
		Handler: "processPaidSessionsHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(processPaidSessionsJob, processPaidSessionsJob.Args["duration"].(time.Duration))

	return nil
}

var claimUnusedPRLsHandler = func(args worker.Args) error {
	thresholdTime := time.Now().Add(-3 * time.Hour) // consider a transaction timed out if it takes more than 3 hours
	ClaimUnusedPRLs(EthWrapper, thresholdTime)

	claimUnusedPRLsJob := worker.Job{
		Queue:   "default",
		Handler: "claimUnusedPRLsHandler",
		Args:    args,
	}
	OysterWorker.PerformIn(claimUnusedPRLsJob, claimUnusedPRLsJob.Args["duration"].(time.Duration))

	return nil
}
