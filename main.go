package main

import (
	"fmt"
	"os"
)

func main() {
	ci, err := getCloudInfo()
	if err != nil {
		fmt.Printf("couldn't collect info about cloud availability zones: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("got %d compute zones\n", len(ci.ComputeZones))
	fmt.Printf("got %d volume zones\n", len(ci.VolumeZones))
}
