package jobs

import (
	"fmt"
	raven "github.com/getsentry/raven-go"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/oysterprotocol/brokernode/services"
	"reflect"
	"runtime"
	"time"
)

const (
	BundleSize = 100
	Duration   = "duration"
)

var OysterWorker = worker.NewSimple()

var (
	IotaWrapper = services.IotaWrapper
	EthWrapper  = services.EthWrapper
)

func init() {
	registerHandlers(OysterWorker)

	doWork(OysterWorker)
}

func registerHandlers(oysterWorker *worker.Simple) {
	logIfError(oysterWorker.Register(getHandlerName(flushOldWebnodesHandler), flushOldWebnodesHandler))
	logIfError(oysterWorker.Register(getHandlerName(processUnassignedChunksHandler), processUnassignedChunksHandler))
	logIfError(oysterWorker.Register(getHandlerName(purgeCompletedSessionsHandler), purgeCompletedSessionsHandler))
	logIfError(oysterWorker.Register(getHandlerName(verifyDataMapsHandler), verifyDataMapsHandler))
	logIfError(oysterWorker.Register(getHandlerName(updateTimedOutDataMapsHandler), updateTimedOutDataMapsHandler))
	logIfError(oysterWorker.Register(getHandlerName(processPaidSessionsHandler), processPaidSessionsHandler))
	logIfError(oysterWorker.Register(getHandlerName(claimUnusedPRLsHandler), claimUnusedPRLsHandler))
	logIfError(oysterWorker.Register(getHandlerName(removeUnpaidUploadSessionHandler), removeUnpaidUploadSessionHandler))
}

func doWork(oysterWorker *worker.Simple) {
	oysterWorkerPerformIn(flushOldWebnodesHandler,
		worker.Args{
			Duration: 5 * time.Minute,
		})

	oysterWorkerPerformIn(processUnassignedChunksHandler,
		worker.Args{
			Duration: time.Duration(services.GetProcessingFrequency()) * time.Second,
		})

	oysterWorkerPerformIn(purgeCompletedSessionsHandler,
		worker.Args{
			Duration: 60 * time.Second,
		})

	oysterWorkerPerformIn(verifyDataMapsHandler,
		worker.Args{
			Duration: 30 * time.Second,
		})

	oysterWorkerPerformIn(updateTimedOutDataMapsHandler,
		worker.Args{
			Duration: 60 * time.Second,
		})

	oysterWorkerPerformIn(processPaidSessionsHandler,
		worker.Args{
			Duration: 30 * time.Second,
		})

	oysterWorkerPerformIn(claimUnusedPRLsHandler,
		worker.Args{
			Duration: 10 * time.Minute,
		})

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler,
		worker.Args{
			Duration: 24 * time.Hour,
		})
}

func flushOldWebnodesHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-20 * time.Minute) // webnodes older than 20 minutes get deleted
	FlushOldWebNodes(thresholdTime)

	oysterWorkerPerformIn(flushOldWebnodesHandler, args)
	return nil
}

func processUnassignedChunksHandler(args worker.Args) error {
	ProcessUnassignedChunks(IotaWrapper)

	oysterWorkerPerformIn(processUnassignedChunksHandler, args)
	return nil
}

func purgeCompletedSessionsHandler(args worker.Args) error {
	PurgeCompletedSessions()

	oysterWorkerPerformIn(purgeCompletedSessionsHandler, args)
	return nil
}

func verifyDataMapsHandler(args worker.Args) error {
	VerifyDataMaps(IotaWrapper)

	oysterWorkerPerformIn(verifyDataMapsHandler, args)
	return nil
}

func updateTimedOutDataMapsHandler(args worker.Args) error {
	UpdateTimeOutDataMaps(time.Now().Add(-1 * time.Minute))

	oysterWorkerPerformIn(updateTimedOutDataMapsHandler, args)
	return nil
}

func processPaidSessionsHandler(args worker.Args) error {
	ProcessPaidSessions()

	oysterWorkerPerformIn(processPaidSessionsHandler, args)
	return nil
}

func claimUnusedPRLsHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-3 * time.Hour) // consider a transaction timed out if it takes more than 3 hours
	ClaimUnusedPRLs(EthWrapper, thresholdTime)

	oysterWorkerPerformIn(claimUnusedPRLsHandler, args)
	return nil
}

func removeUnpaidUploadSessionHandler(args worker.Args) error {
	RemoveUpaidUploadSession()

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler, args)
	return nil
}

func oysterWorkerPerformIn(handler worker.Handler, args worker.Args) {
	job := worker.Job{
		Queue:   "default",
		Handler: getHandlerName(handler),
		Args:    args,
	}
	logIfError(OysterWorker.PerformIn(job, args[Duration].(time.Duration)))
}

func logIfError(err error) {
	if err != nil {
		fmt.Println(err)
		raven.CaptureError(err, nil)
	}
}

// Return the name of the handler in full path.
// ex: github.com/oysterprotocol/brokernode/jobs.flushOldWebnodesHandler
func getHandlerName(i worker.Handler) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
