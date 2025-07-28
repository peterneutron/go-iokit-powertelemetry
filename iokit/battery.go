// Package iokit provides direct access to macOS IOKit power and battery telemetry.
package iokit

// BatteryInfo holds a comprehensive snapshot of all data points retrieved
// from the AppleSmartBattery service in IOKit.
type BatteryInfo struct {
	// IsCharging indicates if the battery is currently charging.
	IsCharging bool
	// IsConnected indicates if an external power source is connected.
	IsConnected bool
	// FullyCharged indicates if the battery is at 100% and not drawing charge.
	FullyCharged bool

	// Health & Capacity - these values are the core of battery health assessment.
	// All capacity values are in milliamp-hours (mAh).
	Health           Health
	Capacity         Capacity
	Charge           Charge
	Temperature      Temperature
	Power            Power
	Hardware         Hardware
	Adapter          Adapter
	PowerSourceInput PowerSourceInput
}

// Health contains metrics related to the battery's long-term condition.
type Health struct {
	// CycleCount is the number of charge/discharge cycles the battery has undergone.
	CycleCount int
}

// Capacity stores the various milliamp-hour (mAh) capacity metrics.
type Capacity struct {
	// DesignCapacity is the "as-new" capacity specified by the manufacturer. This value does not change.
	DesignCapacity int
	// MaxCapacity is the battery's current maximum capacity, as estimated by the BMS.
	// This value degrades over time. It corresponds to IOKit's `AppleRawMaxCapacity`.
	MaxCapacity int
	// NominalCapacity is a smoothed, less volatile capacity value. It is likely used by macOS
	// for the "official" health percentage displayed in System Settings.
	// It corresponds to IOKit's `NominalChargeCapacity`.
	NominalCapacity int
}

// Charge contains the live state of the battery's charge.
type Charge struct {
	// CurrentCapacity is the current charge level in mAh.
	CurrentCapacity int
	// TimeToEmpty is the estimated minutes until the battery is empty (if discharging).
	TimeToEmpty int
	// TimeToFull is the estimated minutes until the battery is full (if charging).
	TimeToFull int
}

// Temperature contains temperature readings in Celsius.
type Temperature struct {
	// Battery is the primary temperature of the battery pack.
	Battery float64
	// IndividualCellVoltages contains the voltage of each cell block in millivolts (mV).
	IndividualCellVoltages []int
}

// Power contains live electrical data for the battery.
type Power struct {
	// Voltage is the current battery voltage in Volts.
	Voltage float64
	// Amperage is the current flowing into/out of the battery in Amps.
	// A negative value indicates the battery is discharging.
	Amperage float64
}

// Hardware contains identifiers for the battery hardware.
type Hardware struct {
	// SerialNumber is the battery's unique serial number.
	SerialNumber string
	// DeviceName is the model name of the battery management system (e.g., "bq40z651").
	DeviceName string
}

// Adapter contains information about the connected power adapter.
type Adapter struct {
	// Watts is the negotiated power rating of the adapter in Watts.
	Watts int
	// Voltage is the negotiated voltage in Volts.
	Voltage float64
	// Amperage is the maximum current the adapter can provide at the negotiated voltage, in Amps.
	Amperage float64
	// Description is a system-provided description (e.g., "pd charger").
	Description string
}

// PowerSourceInput contains live electrical data for the power being drawn
// from the connected adapter.
type PowerSourceInput struct {
	// Voltage is the actual voltage being supplied by the adapter in Volts.
	Voltage float64
	// Amperage is the actual current being drawn by the system in Amps.
	Amperage float64
}
