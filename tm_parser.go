package main

import (
	"io/ioutil"
	"time"

	"github.com/DHowett/go-plist"
)

//Root represents the root of a timemachine.plist file
type Root struct {
	LocalBackups bool `plist:"MobileBackups"`
	AutoBackup   bool
	Destinations []Destination
}

//Destination represents a single destination in the timemachine.plist file
type Destination struct {
	ID            string `plist:"DestinationID"`
	BytesUsed     int64
	SnapshotDates []time.Time
}

func parseTimemachinePlist(path string) (*Root, error) {
	timemachinePlist, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var root Root
	_, err = plist.Unmarshal(timemachinePlist, &root)
	if err != nil {
		return nil, err
	}
	return &root, nil
}
