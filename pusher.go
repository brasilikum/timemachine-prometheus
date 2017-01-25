package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const namespace = "timemachine"

var (
	localBackups = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "local_backups_enabled_bool",
		Help:      "Bool indicating if local backups (also known as mobile backups) are enabled",
		Namespace: namespace,
	})

	autoBackup = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "auto_backups_enabled_bool",
		Help:      "Bool indicating if automatic backups are enabled",
		Namespace: namespace,
	})

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

	lastGathered = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "last_gathered_timestamp_seconds",
		Help:      "Timesamp of the last sucessful gatherig of this metrics",
		Namespace: namespace,
	})
)

var (
	gatherEvery     = flag.Duration("every", time.Minute*10, "time between gathering loops")
	pushgatewayAddr = flag.String("pushgateway.addr", "http://localhost:9091", "address of the pushgateway")
)

func main() {
	flag.Parse()

	registry := prometheus.NewRegistry()
	registry.MustRegister(localBackups, autoBackup, totalSnapshots, bytesUsed, snapshotTime, lastGathered)

	var destination, err = getDesinationAlias()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(destination)

	for {
		tmRoot, err := parseTimemachinePlist("/Library/Preferences/com.apple.TimeMachine.plist")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		//set global metrics
		if tmRoot.AutoBackup {
			autoBackup.Set(1)
		} else {
			autoBackup.Set(0)
		}
		if tmRoot.LocalBackups {
			localBackups.Set(1)
		} else {
			localBackups.Set(0)
		}

		//set destination dependent metrics
		for _, destination := range tmRoot.Destinations {
			labels := prometheus.Labels{"destination_id": destination.ID}
			totalSnapshots.With(labels).Set(float64(len(destination.SnapshotDates)))
			bytesUsed.With(labels).Set(float64(destination.BytesUsed))

			latestSnapshot := destination.SnapshotDates[len(destination.SnapshotDates)-1]
			snapshotTime.With(labels).Set(float64(latestSnapshot.Unix()))
		}

		hostname, err := os.Hostname()
		if err != nil {
			fmt.Println("Could not get hostname:", err)
			os.Exit(1)
		}
		lastGathered.SetToCurrentTime()
		if err := push.FromGatherer(
			namespace,
			prometheus.Labels{"instance": hostname},
			*pushgatewayAddr,
			registry,
		); err != nil {
			fmt.Println("Could not push to Pushgateway:", err)
		}
		time.Sleep(*gatherEvery)
	}
}
