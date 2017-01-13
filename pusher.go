package main

import (
	"fmt"
	"os"
	"time"

	"math/rand"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const namespace = "timemachine"

var (
	totalSnapshots = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "snapshots_total",
		Help:      "Counter of snapshots per destination",
		Namespace: namespace,
	}, []string{"destination_id"})

	bytesUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "disk_used_bytes",
		Help:      "Counter of bytes used per destination",
		Namespace: namespace,
	}, []string{"destination_id"})

	snapshotTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "last_snapshot_timestamp_seconds",
		Help:      "The timestamp of the last completed snapshot",
		Namespace: namespace,
	}, []string{"destination_id"})
)

func main() {
	registry := prometheus.NewRegistry()
	registry.MustRegister(totalSnapshots, bytesUsed, snapshotTime)

	for {
		totalSnapshots.With(prometheus.Labels{"destination_id": "abc"}).Inc()
		bytesUsed.With(prometheus.Labels{"destination_id": "abc"}).Set(rand.Float64())
		snapshotTime.With(prometheus.Labels{"destination_id": "abc"}).SetToCurrentTime()
		snapshotTime.With(prometheus.Labels{"destination_id": "cde"}).SetToCurrentTime()

		hostname, err := os.Hostname()
		if err != nil {
			fmt.Println("Could not get hostname:", err)
			os.Exit(1)
		}
		// AddFromGatherer is used here rather than FromGatherer to not delete a
		// previously pushed success timestamp in case of a failure of this
		// backup.
		if err := push.AddFromGatherer(
			namespace,
			prometheus.Labels{"instance": hostname},
			"http://localhost:9091",
			registry,
		); err != nil {
			fmt.Println("Could not push to Pushgateway:", err)
		}
		time.Sleep(time.Second * 3)
	}
}
