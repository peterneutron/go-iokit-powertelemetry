# Go IOKit Power Telemetry

⚠️ Pre-Release Software Notice ⚠️

This library is in its initial development phase (v0.x.x). The API is not yet stable and is subject to breaking changes in future releases. Please use with caution and consider pinning to a specific version in your project.

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

### How Health and Power are Calculated

The `Calculations` struct provides several derived metrics for convenience. Health percentages are rounded to the nearest whole number, and wattages are truncated to two decimal places.

-   **`Calculations.HealthByMaxCapacity` (int)**
    This represents the "true" physical health of the battery's chemistry, calculated as `(Battery.MaxCapacity / Battery.DesignCapacity) * 100`. It directly reflects the current maximum charge the battery can hold compared to when it was new.

-   **`Calculations.HealthByNominalCapacity` (int)**
    This provides a more stable health percentage that is less affected by recent charge/discharge cycles, making it a reliable historical indicator. It is calculated as `(Battery.NominalCapacity / Battery.DesignCapacity) * 100`.

-   **`Calculations.ConditionAdjustedHealth` (int)**
    This is our reverse-engineered metric that attempts to estimate the health percentage shown by macOS. It starts with the stable `HealthByNominalCapacity` and applies a bonus or penalty based on the **voltage drift** between the battery's internal cell blocks (`Battery.IndividualCellVoltages`). A well-balanced battery receives a health bonus, while an imbalanced battery is penalized.

-   **`Calculations.ACPower` (float64)**
    The total power in Watts currently being drawn from the AC adapter. This value will be zero if the adapter is not connected. It is calculated as `Adapter.InputVoltage * Adapter.InputAmperage`.

-   **`Calculations.BatteryPower` (float64)**
    The power in Watts currently flowing into or out of the battery. A **positive** value indicates the battery is charging, while a **negative** value indicates it is discharging. It is calculated as `Battery.Voltage * Battery.Amperage`.

-   **`Calculations.SystemPower` (float64)**
    An estimate of the power in Watts being consumed by the system hardware (CPU, display, etc.). This value represents the net power usage of the machine itself, calculated as `ACPower - BatteryPower`.

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
  "State": {
    "IsCharging": true,
    "IsConnected": true,
    "FullyCharged": false
  },
  "Battery": {
    "SerialNumber": "F8YH5900JDP00000E7",
    "DeviceName": "bq40z651",
    "CycleCount": 180,
    "DesignCapacity": 8579,
    "MaxCapacity": 7697,
    "NominalCapacity": 7941,
    "CurrentCapacity": 3790,
    "TimeToEmpty": 65535,
    "TimeToFull": 124,
    "Temperature": 30.41,
    "Voltage": 11.932,
    "Amperage": 4.437,
    "IndividualCellVoltages": [
      3979,
      3977,
      3976
    ]
  },
  "Adapter": {
    "Description": "pd charger",
    "MaxWatts": 65,
    "MaxVoltage": 20,
    "MaxAmperage": 3.25,
    "InputVoltage": 19.517,
    "InputAmperage": 3.213
  },
  "Calculations": {
    "HealthByMaxCapacity": 90,
    "HealthByNominalCapacity": 93,
    "ConditionAdjustedHealth": 95,
    "ACPower": 62.7,
    "BatteryPower": 52.94,
    "SystemPower": 9.76
  }
}
```