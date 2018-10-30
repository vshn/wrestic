package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// snapshot models a restic a single snapshot from the
// snapshots --json subcommand.
type snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Tree     string    `json:"tree"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	UID      int       `json:"uid"`
	Gid      int       `json:"gid"`
	Tags     []string  `json:"tags"`
}

func listSnapshots() ([]snapshot, error) {
	args := []string{"snapshots", "--json", "-q"}
	var timeout int
	var converr error

	if timeout, converr = strconv.Atoi(os.Getenv(listTimeoutEnv)); converr != nil {
		timeout = 30
	}

	done := make(chan bool)
	stdout := make([]string, 0)
	stderr := make([]string, 0)
	go func() {
		stdout, stderr = genericCommand(args, commandOptions{print: false})
		done <- true
	}()
	fmt.Printf("Listing snapshots, timeout: %v\n", timeout)
	select {
	case <-done:
		if len(stderr) > 0 && strings.Contains(stderr[1], "following location?") {
			commandError = nil
			return nil, errors.New("Not initialised yet")
		}
		snapList := make([]snapshot, 0)
		output := strings.Join(stdout, "\n")
		err := json.Unmarshal([]byte(output), &snapList)
		if err != nil {
			fmt.Printf("Error listing snapshots\n%v\n%v", err, strings.Join(stderr, "\n"))
			return nil, err
		}
		availableSnapshots := len(snapList)
		fmt.Printf("%v command:\n%v Snapshots\n", args[0], availableSnapshots)
		metrics.AvailableSnapshots.Set(float64(availableSnapshots))
		metrics.Update(metrics.AvailableSnapshots)
		return snapList, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		commandError = errors.New("connection timed out")
		return nil, commandError
	}
}