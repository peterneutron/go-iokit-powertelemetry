# Go IOKit Power Telemetry

[![Go Reference](https://pkg.go.dev/badge/github.com/peterneutron/go-iokit-powertelemetry.svg)](https://pkg.go.dev/github.com/peterneutron/go-iokit-powertelemetry)

A dependency-free Go library for accessing raw macOS power and battery telemetry directly from the underlying IOKit services.

## Features

*   **Direct Hardware Access**: Get data directly from the `AppleSmartBattery` kernel service.
*   **Zero Dependencies**: A lightweight solution with no external Go modules required.
*   **Comprehensive Data**: Provides a detailed snapshot of capacity, charge state, cycle count, temperatures, cell voltages, and adapter information.
*   **Advanced Health Metrics**: Goes beyond raw data to provide calculated, easy-to-understand battery health percentages.

## Installation

```bash
go get github.com/peterneutron/go-iokit-powertelemetry
```

## Quick Start

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

	// Access raw data and calculated health metrics
	fmt.Printf("Cycle Count: %d\n", info.Health.CycleCount)
	fmt.Printf("Health by Max Capacity: %d%%\n", info.Calculations.HealthByMaxCapacity)
	fmt.Printf("Condition-Adjusted Health: %d%%\n", info.Calculations.ConditionAdjustedHealth)
}
```

## Understanding the Data

The library provides both raw capacity metrics and derived health calculations.

### Core Capacity Metrics

These values are read directly from IOKit and form the basis for all health calculations.

| Field                     | Data Source               | Description                                                                                             |
| ------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------- |
| `Capacity.DesignCapacity` | `DesignCapacity`          | The factory-rated capacity of a brand-new battery. This is the baseline for all health calculations.                      |
| `Capacity.MaxCapacity`    | `AppleRawMaxCapacity`     | The current, real-world maximum capacity. This reflects aging and any temporary limits imposed by macOS (e.g., Optimized Battery Charging) and can fluctuate.     |
| `Capacity.NominalCapacity`| `NominalChargeCapacity`   | A more stable, smoothed historical capacity value that is less prone to short-term fluctuations.        |

### How Health is Calculated

The `Calculations` struct provides several health percentages, rounded to the nearest whole number.

> #### `Calculations.HealthByMaxCapacity`
>
> **Formula:** `(MaxCapacity / DesignCapacity) * 100`

> #### `Calculations.HealthByNominalCapacity`
>
> **Formula:** `(NominalCapacity / DesignCapacity) * 100`

> #### `Calculations.ConditionAdjustedHealth`
>
> **Formula:** `HealthByNominalCapacity + Condition Modifier`
>
> This is an experimental metric that attempts to estimate the health percentage shown by macOS. It starts with the stable `HealthByNominalCapacity` and applies a bonus or penalty based on the **voltage drift** between the battery's internal cell blocks (`Battery.IndividualCellVoltages`). A well-balanced battery with low drift receives a health bonus, while an imbalanced battery is penalized, providing a more holistic view of its condition.

## Full Data Example

You can run the included example to see all available data fields.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/peterneutron/go-iokit-powertelemetry.git
    cd go-iokit-powertelemetry
    ```

2.  **Run the example:**
    ```bash
    go run ./examples/simple-dump
    ```

**Example Output (reflects new structure):**

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
  "Battery": {
    "Temperature": 30.6,
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
    "HealthByMaxCapacity": 90,
    "HealthByNominalCapacity": 93,
    "ConditionAdjustedHealth": 95
  }
}
```