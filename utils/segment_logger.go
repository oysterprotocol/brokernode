package oyster_utils

import (
	"github.com/iotaledger/giota"
	"gopkg.in/segmentio/analytics-go.v3"
	"net"
	"os"
	"time"
)

var (
	segmentWriteKey string
	SegmentClient   analytics.Client
)

func init() {
	//// Setup Segment
	segmentWriteKey = os.Getenv("SEGMENT_WRITE_KEY")
	if segmentWriteKey != "" {
		SegmentClient = analytics.New(segmentWriteKey)
	}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func MapTransactionsToAddrs(txs []giota.Transaction) (addrs []giota.Address) {

	addrs = make([]giota.Address, 0, len(txs))

	for _, tx := range txs {
		addrs = append(addrs, tx.Address)
	}
	return
}

func TimeTrack(start time.Time, name string, properties analytics.Properties) {
	if segmentWriteKey != "" {
		elapsed := time.Since(start).Seconds()

		go SegmentClient.Enqueue(analytics.Track{
			Event:  name,
			UserId: GetLocalIP(),
			Properties: properties.
				Set("time_elapsed", elapsed),
		})
	}
}
