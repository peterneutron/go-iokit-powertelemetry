# Go IOKit Power Telemetry

[![Go Reference](https://pkg.go.dev/badge/github.com/peterneutron/go-iokit-powertelemetry.svg)](https://pkg.go.dev/github.com/peterneutron/go-iokit-powertelemetry)

A dependency-free Go library for directly accessing macOS power and battery telemetry using IOKit.

This library bypasses command-line tools like `system_profiler` to get raw, unformatted data directly from the `AppleSmartBattery` kernel service.

## Installation

```bash
go get github.com/peterneutron/go-iokit-powertelemetry/iokit
```

## Usage

Here is a minimal example of how to import and use the library in your own project.

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/peterneutron/go-iokit-powertelemetry/iokit"
)

func main() {
	info, err := iokit.GetBatteryInfo()
	if err != nil {
		log.Fatalf("Error getting battery info: %v", err)
	}

	// You can now access any value from the info struct.
	// For example, to get the cycle count and estimated health:
	fmt.Printf("Cycle Count: %d\n", info.Health.CycleCount)
	fmt.Printf("Estimated Official Health: %.2f%%\n", info.Calculations.EstimatedOfficialHealth)
}
```

### Running the Full Example

The repository includes an example that prints all retrieved data as a JSON object. To run it:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/peterneutron/go-iokit-powertelemetry.git
    cd go-iokit-powertelemetry
    ```

2.  **Run the simple dump example:**
    ```bash
    go run ./examples/simple-dump
    ```

**Example Output:**

```json
{
  "IsCharging": false,
  "IsConnected": false,
  "FullyCharged": false,
  "Health": {
    "CycleCount": 180
  },
  "Capacity": {
    "DesignCapacity": 8579,
    "MaxCapacity": 7701,
    "NominalCapacity": 7945
  },
  "Charge": {
    "CurrentCapacity": 2745,
    "TimeToEmpty": 178,
    "TimeToFull": 65535
  },
  "Temperature": {
    "Battery": 30.6,
    "IndividualCellVoltages": [
      3783,
      3785,
      3784
    ]
  },
  "Power": {
    "Voltage": 11.353,
    "Amperage": -0.92
  },
  "Hardware": {
    "SerialNumber": "xxxxxxxxxxxxxxxxxx",
    "DeviceName": "xxxxxxxx"
  },
  "Adapter": {
    "Watts": 65,
    "Voltage": 20,
    "Amperage": 3.25,
    "Description": "pd charger"
  },
  "PowerSourceInput": {
    "Voltage": 20.188,
    "Amperage": 0.005
  },
  "Calculations": {
    "HealthPercentage": 89.765706958853,
    "NominalHealthPercentage": 92.60986128919454,
    "EstimatedOfficialHealth": 95.10986128919454
  }
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
