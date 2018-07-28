package jobs

import (
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/oysterprotocol/brokernode/services"
	"github.com/oysterprotocol/brokernode/utils"
)

const (
	/*BundleSize is the number of transactions per iota broadcast*/
	BundleSize = 100
	/*Duration is for one of our tags for error logging*/
	Duration = "duration"
	/*SecondsDelayForETHPolling is how long to wait between polling attempts for ethereum transactions*/
	SecondsDelayForETHPolling = 1 * 60
)

var OysterWorker = worker.NewSimple()

var (
	IotaWrapper       = services.IotaWrapper
	EthWrapper        = services.EthWrapper
	PrometheusWrapper = services.PrometheusWrapper
)

func init() {
	if oyster_utils.IsInUnitTest() {
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
	oysterWorker.Register(getHandlerName(claimTreasureForWebnodeHandler), claimTreasureForWebnodeHandler)
	oysterWorker.Register(getHandlerName(removeUnpaidUploadSessionHandler), removeUnpaidUploadSessionHandler)
	oysterWorker.Register(getHandlerName(updateMsgStatusHandler), updateMsgStatusHandler)

	oysterWorker.Register(getHandlerName(buryTreasureAddressesHandler), buryTreasureAddressesHandler)
	oysterWorker.Register(getHandlerName(claimUnusedPRLsHandler), claimUnusedPRLsHandler)
	oysterWorker.Register(getHandlerName(checkAlphaPaymentsHandler), checkAlphaPaymentsHandler)
	oysterWorker.Register(getHandlerName(checkBetaPaymentsHandler), checkBetaPaymentsHandler)
	oysterWorker.Register(getHandlerName(storeCompletedGenesisHashesHandler), storeCompletedGenesisHashesHandler)

	if services.IsKvStoreEnabled() {
		oysterWorker.Register(getHandlerName(badgerDbGcHandler), badgerDbGcHandler)
	}
}

func doWork(oysterWorker *worker.Simple) {
	oysterWorkerPerformIn(flushOldWebnodesHandler,
		worker.Args{Duration: 5 * time.Minute})

	oysterWorkerPerformIn(processUnassignedChunksHandler,
		worker.Args{Duration: time.Duration(services.GetProcessingFrequency()) * time.Second})

	oysterWorkerPerformIn(purgeCompletedSessionsHandler,
		worker.Args{Duration: 10 * time.Minute})

	oysterWorkerPerformIn(verifyDataMapsHandler,
		worker.Args{Duration: 30 * time.Second})

	oysterWorkerPerformIn(updateTimedOutDataMapsHandler,
		worker.Args{Duration: 60 * time.Second})

	oysterWorkerPerformIn(processPaidSessionsHandler,
		worker.Args{Duration: 20 * time.Second})

	oysterWorkerPerformIn(claimTreasureForWebnodeHandler,
		worker.Args{Duration: 2 * time.Minute})

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler,
		worker.Args{Duration: 24 * time.Hour})

	oysterWorkerPerformIn(updateMsgStatusHandler,
		worker.Args{Duration: 15 * time.Second})

	oysterWorkerPerformIn(badgerDbGcHandler,
		worker.Args{Duration: 10 * time.Minute})

	if oyster_utils.BrokerMode != oyster_utils.ProdMode {
		oysterWorkerPerformIn(storeCompletedGenesisHashesHandler,
			worker.Args{Duration: 1 * time.Minute})
	}

	if oyster_utils.BrokerMode == oyster_utils.ProdMode && oyster_utils.PaymentMode == oyster_utils.UserIsPaying {
		oysterWorkerPerformIn(buryTreasureAddressesHandler,
			worker.Args{Duration: 2 * time.Minute})

		oysterWorkerPerformIn(claimUnusedPRLsHandler,
			worker.Args{Duration: 10 * time.Minute})

		oysterWorkerPerformIn(checkAlphaPaymentsHandler,
			worker.Args{Duration: 10 * time.Second})

		oysterWorkerPerformIn(checkBetaPaymentsHandler,
			worker.Args{Duration: 70 * time.Second})
	}
}

func flushOldWebnodesHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-20 * time.Minute) // webnodes older than 20 minutes get deleted
	FlushOldWebNodes(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(flushOldWebnodesHandler, args)
	return nil
}

func processUnassignedChunksHandler(args worker.Args) error {

	if os.Getenv("TANGLE_MAINTENANCE") != "true" {
		ProcessUnassignedChunks(IotaWrapper, PrometheusWrapper)
	}

	oysterWorkerPerformIn(processUnassignedChunksHandler, args)
	return nil
}

func purgeCompletedSessionsHandler(args worker.Args) error {
	PurgeCompletedSessions(PrometheusWrapper)

	oysterWorkerPerformIn(purgeCompletedSessionsHandler, args)
	return nil
}

func verifyDataMapsHandler(args worker.Args) error {
	if os.Getenv("TANGLE_MAINTENANCE") != "true" {
		VerifyDataMaps(IotaWrapper, PrometheusWrapper)
	}

	oysterWorkerPerformIn(verifyDataMapsHandler, args)
	return nil
}

func updateTimedOutDataMapsHandler(args worker.Args) error {
	UpdateTimeOutDataMaps(time.Now().Add(-2*time.Minute), PrometheusWrapper)

	oysterWorkerPerformIn(updateTimedOutDataMapsHandler, args)
	return nil
}

func processPaidSessionsHandler(args worker.Args) error {
	ProcessPaidSessions(PrometheusWrapper)

	oysterWorkerPerformIn(processPaidSessionsHandler, args)
	return nil
}

func updateMsgStatusHandler(args worker.Args) error {
	UpdateMsgStatus(PrometheusWrapper)

	oysterWorkerPerformIn(updateMsgStatusHandler, args)
	return nil
}

func buryTreasureAddressesHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-18 * time.Hour) // consider a transaction timed out after 18 hours
	BuryTreasureAddresses(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(buryTreasureAddressesHandler, args)
	return nil
}

func claimTreasureForWebnodeHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-18 * time.Hour) // consider a transaction timed out after 18 hours
	ClaimTreasureForWebnode(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(claimTreasureForWebnodeHandler, args)
	return nil
}

func claimUnusedPRLsHandler(args worker.Args) error {
	thresholdTime := time.Now().Add(-18 * time.Hour) // consider a transaction timed out after 18 hours
	ClaimUnusedPRLs(thresholdTime, PrometheusWrapper)

	oysterWorkerPerformIn(claimUnusedPRLsHandler, args)
	return nil
}

func removeUnpaidUploadSessionHandler(args worker.Args) error {
	RemoveUnpaidUploadSession(PrometheusWrapper)

	oysterWorkerPerformIn(removeUnpaidUploadSessionHandler, args)
	return nil
}

func checkAlphaPaymentsHandler(args worker.Args) error {
	CheckAlphaPayments(PrometheusWrapper)

	oysterWorkerPerformIn(checkAlphaPaymentsHandler, args)
	return nil
}

func checkBetaPaymentsHandler(args worker.Args) error {
	durationToWaitBeforeTimingOut := time.Duration(-6 * time.Hour)
	// consider a beta transaction timed out after 6 hours if on alpha
	// consider a beta transaction timed out after 18 hours if on beta
	CheckBetaPayments(durationToWaitBeforeTimingOut, PrometheusWrapper)

	oysterWorkerPerformIn(checkBetaPaymentsHandler, args)
	return nil
}

func storeCompletedGenesisHashesHandler(args worker.Args) error {
	StoreCompletedGenesisHashes(PrometheusWrapper)

	oysterWorkerPerformIn(storeCompletedGenesisHashesHandler, args)
	return nil
}

func badgerDbGcHandler(args worker.Args) error {
	BadgerDbGc()

	oysterWorkerPerformIn(badgerDbGcHandler, args)
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
