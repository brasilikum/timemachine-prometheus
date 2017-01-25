package main

import (
	"encoding/xml"
	"os/exec"

	"github.com/pkg/errors"
)

type DestinationInfo struct{}

func getDesinationAlias() (*DestinationInfo, error) {
	var cmd = *exec.Command("tmutil", "destinationinfo", "-X")
	var cmdOutput, err = cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not get cmdOutput")
	}
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "could not run tmutil")
	}
	var destInfo *DestinationInfo
	err = xml.NewDecoder(cmdOutput).Decode(destInfo)
	return destInfo, err
}
