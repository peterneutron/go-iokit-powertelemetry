package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/peterneutron/go-iokit-powertelemetry/iokit"
)

func main() {
	info, err := iokit.GetBatteryInfo()
	if err != nil {
		log.Fatalf("Error getting battery info: %v", err)
	}

	// Print the data as a nicely formatted JSON object.
	// This is a great way to see everything at once.
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
