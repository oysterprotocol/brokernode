package jobs

import (
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

const (
	BundleSize = 300
	Duration   = "duration"
)

var OysterWorker = worker.NewSimple()

var (
	IotaWrapper       = services.IotaWrapper
	EthWrapper        = services.EthWrapper
	PrometheusWrapper = services.PrometheusWrapper
)

func init() {
	if enabled, err := strconv.ParseBool(os.Getenv("JOB_RUNNER")); err == nil && !enabled {
		return
	}

	registerHandlers(OysterWorker)
	doWork(OysterWorker)
}

func registerHandlers(oysterWorker *worker.Simple) {
	oysterWorker.Register(getHandlerName(flushOldWebnodesHandler), flushOldWebnodesHandler)
	oysterWorker.Register(getHandlerName(processUnassignedChunksHandler), processUnassignedChunksHandler)
	oysterWorker.Register(getHandlerName(purgeCompletedSessionsHandler), purgeCompletedSessionsHandler)
	oysterWorker.Register(getHandlerName(verifyDataMapsHandler), verifyDataMapsHandler)
	oysterWorker.Register(getHandlerName(updateTimedOutDataMapsHandler), updateTimedOutDataMapsHandler)
	oysterWorker.Register(getHandlerName(processPaidSessionsHandler), processPaidSessionsHandler)
	oysterWorker.Register(getHandlerName()))
	if os.Getenv("OYSTER_PAYS") == "" {
		oysterWorker.Register(getHandlerName(claimUnusedPRLsHandler), claimUnusedPRLsHandler)
	}
	oysterWorker.Register(getHandlerName(removeUnpaidUploadSessionHandler), removeUnpaidUploadSessionHandler)

	if services.IsKvStoreEnabled() {
		oysterWorker.Register(getHandlerName(badgerDbGcHandler), badgerDbGcHandler)
	}
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
			Duration: 20 * time.Second,
		})

	oysterWorkerPerformIn(claimUnusedPRLsHandler,
		worker.Args{
			Duration: 10 * time.Minute,
		})

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler,
		worker.Args{
			Duration: 24 * time.Hour,
		})
	oysterWorkerPerformIn(badgerDbGcHandler,
		worker.Args{
			Duration: 10 * time.Minute,
		})
}

func flushOldWebnodesHandler(args worker.Args) {
	thresholdTime := time.Now().Add(-20 * time.Minute) // webnodes older than 20 minutes get deleted
	FlushOldWebNodes(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(flushOldWebnodesHandler, args)
}

func processUnassignedChunksHandler(args worker.Args) {
	ProcessUnassignedChunks(IotaWrapper, PrometheusWrapper)

	oysterWorkerPerformIn(processUnassignedChunksHandler, args)
}

func purgeCompletedSessionsHandler(args worker.Args) {
	PurgeCompletedSessions(PrometheusWrapper)

	oysterWorkerPerformIn(purgeCompletedSessionsHandler, args)
}

func verifyDataMapsHandler(args worker.Args) {
	VerifyDataMaps(IotaWrapper, PrometheusWrapper)

	oysterWorkerPerformIn(verifyDataMapsHandler, args)
}

func updateTimedOutDataMapsHandler(args worker.Args) {
	UpdateTimeOutDataMaps(time.Now().Add(-2*time.Minute), PrometheusWrapper)

	oysterWorkerPerformIn(updateTimedOutDataMapsHandler, args)
}

func processPaidSessionsHandler(args worker.Args) {
	thresholdTime := time.Now().Add(-2 * time.Hour) // consider a transaction timed out after 2 hours
	ProcessPaidSessions(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(processPaidSessionsHandler, args)
}

func claimUnusedPRLsHandler(args worker.Args) {
	thresholdTime := time.Now().Add(-3 * time.Hour) // consider a transaction timed out if it takes more than 3 hours
	ClaimUnusedPRLs(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(claimUnusedPRLsHandler, args)
}

func removeUnpaidUploadSessionHandler(args worker.Args) {
	RemoveUnpaidUploadSession(PrometheusWrapper)

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler, args)
}

func badgerDbGcHandler(args worker.Args) {
	BadgerDbGc()

	oysterWorkerPerformIn(badgerDbGcHandler, args)
}

func oysterWorkerPerformIn(handler worker.Handler, args worker.Args) {
	job := worker.Job{
		Queue:   "default",
		Handler: getHandlerName(handler),
		Args:    args,
	}
	oyster_utils.LogIfError(OysterWorker.PerformIn(job, args[Duration].(time.Duration)), nil)
}

// Return the name of the handler in full path.
// ex: github.com/oysterprotocol/brokernode/jobs.flushOldWebnodesHandler
func getHandlerName(i worker.Handler) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
