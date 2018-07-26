package services

import (
	"github.com/dgraph-io/badger/y"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusCollector struct {
	badgerNumReads        *prometheus.Desc
	badgerNumWrites       *prometheus.Desc
	badgerNumBytesRead    *prometheus.Desc
	badgerNumBytesWritten *prometheus.Desc
	badgerNumGet          *prometheus.Desc
	badgerNumPut          *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		badgerNumReads: prometheus.NewDesc("badger_disk_reads_total",
			"", nil, nil,
		),
		badgerNumWrites: prometheus.NewDesc("badger_disk_writes_total",
			"", nil, nil,
		),
		badgerNumBytesRead: prometheus.NewDesc("badger_read_bytes",
			"", nil, nil,
		),
		badgerNumBytesWritten: prometheus.NewDesc("badger_written_bytes",
			"", nil, nil,
		),
		badgerNumGet: prometheus.NewDesc("badger_gets_total",
			"", nil, nil,
		),
		badgerNumPut: prometheus.NewDesc("badger_puts_total",
			"", nil, nil,
		),
	}
}

func (collector *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.badgerNumReads
	ch <- collector.badgerNumWrites
	ch <- collector.badgerNumBytesRead
	ch <- collector.badgerNumBytesWritten
	ch <- collector.badgerNumGet
	ch <- collector.badgerNumPut
}

func (collector *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.badgerNumReads, prometheus.GaugeValue, float64(y.NumReads.Value()))
	ch <- prometheus.MustNewConstMetric(collector.badgerNumWrites, prometheus.GaugeValue, float64(y.NumWrites.Value()))
	ch <- prometheus.MustNewConstMetric(collector.badgerNumBytesRead, prometheus.GaugeValue, float64(y.NumBytesRead.Value()))
	ch <- prometheus.MustNewConstMetric(collector.badgerNumBytesWritten, prometheus.GaugeValue, float64(y.NumBytesWritten.Value()))
	ch <- prometheus.MustNewConstMetric(collector.badgerNumGet, prometheus.GaugeValue, float64(y.NumGets.Value()))
	ch <- prometheus.MustNewConstMetric(collector.badgerNumPut, prometheus.GaugeValue, float64(y.NumPuts.Value()))
}
