package services

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	PrometheusWrapper PrometheusService
)

type PrepareHistogram func(name string, help string, labelNames ...string) (histogram *prometheus.HistogramVec)
type HistogramSeconds func(histogram *prometheus.HistogramVec, startTime time.Time) ()
type Time func() (startTime time.Time)

type PrometheusService struct {
	PrepareHistogram
	HistogramSeconds
	Time
	HistogramTreasuresVerifyAndClaim *prometheus.HistogramVec
	HistogramUploadSessionResourceCreate *prometheus.HistogramVec
	HistogramUploadSessionResourceUpdate *prometheus.HistogramVec
	HistogramUploadSessionResourceCreateBeta *prometheus.HistogramVec
	HistogramUploadSessionResourceGetPaymentStatus *prometheus.HistogramVec
	HistogramWebnodeResourceCreate *prometheus.HistogramVec
	HistogramTransactionBrokernodeResourceCreate *prometheus.HistogramVec
	HistogramTransactionBrokernodeResourceUpdate *prometheus.HistogramVec
	HistogramTransactionGenesisHashResourceCreate *prometheus.HistogramVec
	HistogramTransactionGenesisHashResourceUpdate *prometheus.HistogramVec
}

func init() {
	histogramTreasuresVerifyAndClaim := prepareHistogram("treasures_verify_and_claim_seconds", "HistogramTreasuresVerifyAndClaimSeconds", "code")
	histogramUploadSessionResourceCreate := prepareHistogram("upload_session_resource_create_seconds", "HistogramUploadSessionResourceCreateSeconds", "code")
	histogramUploadSessionResourceUpdate := prepareHistogram("upload_session_resource_update_seconds", "HistogramUploadSessionResourceUpdateSeconds", "code")
	histogramUploadSessionResourceCreateBeta := prepareHistogram("upload_session_resource_create_beta_seconds", "HistogramUploadSessionResourceCreateBetaSeconds", "code")
	histogramUploadSessionResourceGetPaymentStatus := prepareHistogram("upload_session_resource_get_payment_status_seconds", "HistogramUploadSessionResourceGetPaymentStatusSeconds", "code")
	histogramWebnodeResourceCreate := prepareHistogram("webnode_resource_create_seconds", "HistogramWebnodeResourceCreateSeconds", "code")
	histogramTransactionBrokernodeResourceCreate := prepareHistogram("transaction_brokernode_resource_create_seconds", "HistogramTransactionBrokernodeResourceCreateSeconds", "code")
	histogramTransactionBrokernodeResourceUpdate := prepareHistogram("transaction_brokernode_resource_update_seconds", "HistogramTransactionBrokernodeResourceUpdateSeconds", "code")
	histogramTransactionGenesisHashResourceCreate := prepareHistogram("transaction_genesis_hash_resource_create_seconds", "HistogramTransactionGenesisHashResourceCreateSeconds", "code")
	histogramTransactionGenesisHashResourceUpdate := prepareHistogram("transaction_genesis_hash_resource_seconds", "HistogramTransactionGenesisHashResourceUpdateSeconds", "code")


	PrometheusWrapper = PrometheusService{
		PrepareHistogram: prepareHistogram,
		HistogramSeconds: histogramSeconds,
		Time: timeNow,
		HistogramTreasuresVerifyAndClaim: histogramTreasuresVerifyAndClaim,
		HistogramUploadSessionResourceCreate: histogramUploadSessionResourceCreate,
		HistogramUploadSessionResourceUpdate: histogramUploadSessionResourceUpdate,
		HistogramUploadSessionResourceCreateBeta: histogramUploadSessionResourceCreateBeta,
		HistogramUploadSessionResourceGetPaymentStatus: histogramUploadSessionResourceGetPaymentStatus,
		HistogramWebnodeResourceCreate: histogramWebnodeResourceCreate,
		HistogramTransactionBrokernodeResourceCreate: histogramTransactionBrokernodeResourceCreate,
		HistogramTransactionBrokernodeResourceUpdate: histogramTransactionBrokernodeResourceUpdate,
		HistogramTransactionGenesisHashResourceCreate: histogramTransactionGenesisHashResourceCreate,
		HistogramTransactionGenesisHashResourceUpdate: histogramTransactionGenesisHashResourceUpdate,
	}

}

func duration(start time.Time) (duration time.Duration) {
	duration = time.Since(start)
	return duration
}

func timeNow() (start time.Time) {
	start = time.Now()
	return start
}

func prepareHistogram(name string, help string, labelNames ...string) (histogram *prometheus.HistogramVec) {
	histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: name,
		Help: help,
	}, labelNames)

	prometheus.Register(histogram)
	return histogram
}

func histogramSeconds(histogram *prometheus.HistogramVec, start time.Time) () {
	duration := duration(start)
	histogram.WithLabelValues("500").Observe(duration.Seconds())
}
