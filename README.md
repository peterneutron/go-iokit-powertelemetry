# Go IOKit Power Telemetry

[![Go Reference](https://pkg.go.dev/badge/github.com/peterneutron/go-iokit-powertelemetry.svg)](https://pkg.go.dev/github.com/peterneutron/go-iokit-powertelemetry)

A dependency-free Go library for directly accessing macOS power and battery telemetry using IOKit.

This library bypasses command-line tools like `system_profiler` to get raw, unformatted data directly from the `AppleSmartBattery` kernel service.

## Installation

```bash
go get github.com/peterneutron/go-iokit-powertelemetry/iokit
```

## Usage

```go
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

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
```

## Understanding Health Metrics

This library provides several capacity and health values. Understanding the difference is key to interpreting the data correctly.

*   **`Capacity.DesignCapacity`**: The original, "as-new" capacity of the battery in mAh. This value does not change.
*   **`Capacity.MaxCapacity`**: The battery's current, real-world maximum capacity as estimated by the Battery Management System (BMS). This value degrades over time and can fluctuate slightly based on recent charge cycles. It corresponds to IOKit's `AppleRawMaxCapacity`.
*   **`Capacity.NominalCapacity`**: A more stable, smoothed historical capacity value. This is less prone to short-term fluctuations than `MaxCapacity`.

Based on these values, the `Calculations` struct provides several experimental health percentages:

*   **`Calculations.HealthPercentage`**: Calculated as `(MaxCapacity / DesignCapacity)`. This represents the "true" physical health of the battery's chemistry at this moment.
*   **`Calculations.NominalHealthPercentage`**: Calculated as `(NominalCapacity / DesignCapacity)`. This is a more stable health percentage.
*   **`Calculations.EstimatedOfficialHealth`**: Our reverse-engineered formula that attempts to replicate the percentage shown in macOS's System Settings. It uses `NominalHealthPercentage` as a base and applies a bonus or penalty based on the balance of the battery's cell blocks (`IndividualCellVoltages`). A well-balanced battery receives a health bonus.

**Note:** The official percentage shown by Apple is proprietary and not directly exposed by IOKit. `EstimatedOfficialHealth` is a best-effort calculation and is provided for experimental purposes.
