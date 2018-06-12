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
	enabled, err := strconv.ParseBool(os.Getenv("JOB_RUNNER"))
	isEnvDisabled = err == nil && !enabled
	if isEnvDisabled || oyster_utils.IsUnitTest() {
		return
	}

	registerHandlers(OysterWorker)
	doWork(OysterWorker)
}

func registerHandlers(oysterWorker *worker.Simple) {
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(flushOldWebnodesHandler), flushOldWebnodesHandler), nil)
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(processUnassignedChunksHandler), processUnassignedChunksHandler), nil)
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(purgeCompletedSessionsHandler), purgeCompletedSessionsHandler), nil)
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(verifyDataMapsHandler), verifyDataMapsHandler), nil)
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(updateTimedOutDataMapsHandler), updateTimedOutDataMapsHandler), nil)
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(processPaidSessionsHandler), processPaidSessionsHandler), nil)
	if os.Getenv("OYSTER_PAYS") == "" {
		oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(claimUnusedPRLsHandler), claimUnusedPRLsHandler), nil)
	}
	oyster_utils.LogIfError(oysterWorker.Register(getHandlerName(removeUnpaidUploadSessionHandler), removeUnpaidUploadSessionHandler), nil)
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
}

func flushOldWebnodesHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-20 * time.Minute) // webnodes older than 20 minutes get deleted
	FlushOldWebNodes(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(flushOldWebnodesHandler, args)
	return nil
}

func processUnassignedChunksHandler(args worker.Args) error {
	ProcessUnassignedChunks(IotaWrapper, PrometheusWrapper)

	oysterWorkerPerformIn(processUnassignedChunksHandler, args)
	return nil
}

func purgeCompletedSessionsHandler(args worker.Args) error {
	PurgeCompletedSessions(PrometheusWrapper)

	oysterWorkerPerformIn(purgeCompletedSessionsHandler, args)
	return nil
}

func verifyDataMapsHandler(args worker.Args) error {
	VerifyDataMaps(IotaWrapper, PrometheusWrapper)

	oysterWorkerPerformIn(verifyDataMapsHandler, args)
	return nil
}

func updateTimedOutDataMapsHandler(args worker.Args) error {
	UpdateTimeOutDataMaps(time.Now().Add(-2*time.Minute), PrometheusWrapper)

	oysterWorkerPerformIn(updateTimedOutDataMapsHandler, args)
	return nil
}

func processPaidSessionsHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-2 * time.Hour) // consider a transaction timed out after 2 hours
	ProcessPaidSessions(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(processPaidSessionsHandler, args)
	return nil
}

func claimUnusedPRLsHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-3 * time.Hour) // consider a transaction timed out if it takes more than 3 hours
	ClaimUnusedPRLs(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(claimUnusedPRLsHandler, args)
	return nil
}

func removeUnpaidUploadSessionHandler(args worker.Args) error {
	RemoveUnpaidUploadSession(PrometheusWrapper)

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler, args)
	return nil
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
