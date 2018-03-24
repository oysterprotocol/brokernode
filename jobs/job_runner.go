package jobs

import (
	"github.com/oysterprotocol/brokernode/services"
	"github.com/gobuffalo/buffalo/worker"
	"time"
)

var BundleSize = 10

var OysterWorker = worker.NewSimple()

var IotaWrapper = services.IotaWrapper

func init() {
	registerHandlers(OysterWorker)

	doWork(OysterWorker)
}

func registerHandlers(oysterWorker *worker.Simple) {
	oysterWorker.Register("flushOldWebnodesHandler", flushOldWebnodesHandler)
	oysterWorker.Register("processUnassignedChunksHandler", processUnassignedChunksHandler)
	oysterWorker.Register("purgeCompletedSessionsHandler", purgeCompletedSessionsHandler)
	oysterWorker.Register("verifyDataMapsHandler", verifyDataMapsHandler)
}

func doWork(oysterWorker *worker.Simple) {
	flushOldWebnodesJob := worker.Job{
		Queue:   "default",
		Handler: "flushOldWebnodesHandler",
		Args: worker.Args{
			"duration": 1 * time.Minute,
		},
	}

	processUnassignedChunksJob := worker.Job{
		Queue:   "default",
		Handler: "processUnassignedChunksHandler",
		Args: worker.Args{
			"duration": 1 * time.Minute,
		},
	}

	purgeCompletedSessionsJob := worker.Job{
		Queue:   "default",
		Handler: "purgeCompletedSessionsHandler",
		Args: worker.Args{
			"duration": 1 * time.Minute,
		},
	}

	verifyDataMapsJob := worker.Job{
		Queue:   "default",
		Handler: "verifyDataMapsHandler",
		Args: worker.Args{
			"duration": 1 * time.Minute,
		},
	}

	oysterWorker.PerformIn(flushOldWebnodesJob, flushOldWebnodesJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(processUnassignedChunksJob, processUnassignedChunksJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(purgeCompletedSessionsJob, purgeCompletedSessionsJob.Args["duration"].(time.Duration))
	oysterWorker.PerformIn(verifyDataMapsJob, verifyDataMapsJob.Args["duration"].(time.Duration))
}

var flushOldWebnodesHandler = func(args worker.Args) error {
	thresholdTime := time.Now()
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
