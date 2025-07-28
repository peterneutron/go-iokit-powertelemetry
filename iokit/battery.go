// Package iokit provides direct access to macOS IOKit power and battery telemetry.
package iokit

/*
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit

#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>

// C-side struct to hold the raw data. We use this as an intermediary
// to avoid passing complex Go pointers into C.
typedef struct {
    // Top-level booleans
    int is_charging;
    int is_connected;
    int is_fully_charged;

    // Health
    long cycle_count;

    // Capacity (mAh)
    long design_capacity;
    long max_capacity;
    long nominal_capacity;

    // Charge (mAh)
    long current_capacity;
    long time_to_empty;
    long time_to_full;

    // Temperature (°C * 100)
    long temperature;

    // Power (mV, mA)
    long voltage;
    long amperage;

    // Hardware strings
    char serial_number[256];
    char device_name[256];

    // Adapter Info
    long adapter_watts;
    long adapter_voltage;
    long adapter_amperage;
    char adapter_description[256];

    // Power Source Input (mV, mA)
    long source_voltage;
    long source_amperage;

} c_battery_info;

// Helper to safely get a long integer value from a CFDictionary.
// Returns 0 if key is not found or is not a number.
static long get_long_prop(CFDictionaryRef dict, const char *key) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return 0;

    long value = 0;
    CFNumberRef num_ref = (CFNumberRef)CFDictionaryGetValue(dict, key_ref);
    if (num_ref != NULL && CFGetTypeID(num_ref) == CFNumberGetTypeID()) {
        CFNumberGetValue(num_ref, kCFNumberSInt64Type, &value);
    }

    CFRelease(key_ref);
    return value;
}

// Helper to safely get a boolean value from a CFDictionary.
// Returns 0 (false) if key is not found or is not a boolean.
static int get_bool_prop(CFDictionaryRef dict, const char *key) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return 0;

    int value = 0;
    CFBooleanRef bool_ref = (CFBooleanRef)CFDictionaryGetValue(dict, key_ref);
    if (bool_ref != NULL && CFGetTypeID(bool_ref) == CFBooleanGetTypeID()) {
        value = CFBooleanGetValue(bool_ref);
    }

    CFRelease(key_ref);
    return value;
}

// Helper to safely get a string value from a CFDictionary.
static void get_string_prop(CFDictionaryRef dict, const char *key, char *buffer, int buffer_size) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) { buffer[0] = '\0'; return; }

    CFStringRef str_ref = (CFStringRef)CFDictionaryGetValue(dict, key_ref);
    if (str_ref != NULL && CFGetTypeID(str_ref) == CFStringGetTypeID()) {
        CFStringGetCString(str_ref, buffer, buffer_size, kCFStringEncodingUTF8);
    } else {
        buffer[0] = '\0';
    }
    CFRelease(key_ref);
}

// Helper to get a nested dictionary from a parent dictionary.
// Returns NULL if the key doesn't exist or isn't a dictionary.
static CFDictionaryRef get_dict_prop(CFDictionaryRef dict, const char *key) {
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return NULL;

    CFDictionaryRef value = (CFDictionaryRef)CFDictionaryGetValue(dict, key_ref);
    CFRelease(key_ref);

    if (value != NULL && CFGetTypeID(value) == CFDictionaryGetTypeID()) {
        return value;
    }
    return NULL;
}


// The core C function to get all battery properties.
// Returns 0 on success, non-zero on error.
int get_all_battery_info(c_battery_info *info) {
    // Find the AppleSmartBattery service
    CFMutableDictionaryRef matching = IOServiceMatching("AppleSmartBattery");
    if (matching == NULL) return 1;

    io_iterator_t iterator;
    if (IOServiceGetMatchingServices(kIOMainPortDefault, matching, &iterator) != KERN_SUCCESS) {
        // matching is consumed by the call, no need to release it on success
        CFRelease(matching);
        return 2;
    }

    io_service_t battery = IOIteratorNext(iterator);
    IOObjectRelease(iterator);
    if (battery == IO_OBJECT_NULL) return 3;

    // Get the properties of the battery service
    CFMutableDictionaryRef properties = NULL;
    kern_return_t result = IORegistryEntryCreateCFProperties(battery, &properties, kCFAllocatorDefault, 0);
    IOObjectRelease(battery); // Done with the service object
    if (result != KERN_SUCCESS || properties == NULL) return 4;

    // --- Populate the struct using our safe helpers ---

    info->is_charging = get_bool_prop(properties, "IsCharging");
    info->is_connected = get_bool_prop(properties, "ExternalConnected");
    info->is_fully_charged = get_bool_prop(properties, "FullyCharged");

    info->cycle_count = get_long_prop(properties, "CycleCount");

    info->design_capacity = get_long_prop(properties, "DesignCapacity");
    info->max_capacity = get_long_prop(properties, "AppleRawMaxCapacity");
    info->nominal_capacity = get_long_prop(properties, "NominalChargeCapacity");

    info->current_capacity = get_long_prop(properties, "AppleRawCurrentCapacity");
    info->time_to_empty = get_long_prop(properties, "AvgTimeToEmpty");
    info->time_to_full = get_long_prop(properties, "AvgTimeToFull");

    info->temperature = get_long_prop(properties, "Temperature");

    info->voltage = get_long_prop(properties, "Voltage");
    info->amperage = get_long_prop(properties, "Amperage");

    get_string_prop(properties, "Serial", info->serial_number, 256);
    get_string_prop(properties, "DeviceName", info->device_name, 256);

    // Get nested adapter info
    CFDictionaryRef adapter_details = get_dict_prop(properties, "AdapterDetails");
    if (adapter_details) {
        info->adapter_watts = get_long_prop(adapter_details, "Watts");
        info->adapter_voltage = get_long_prop(adapter_details, "AdapterVoltage");
        info->adapter_amperage = get_long_prop(adapter_details, "Current");
        get_string_prop(adapter_details, "Description", info->adapter_description, 256);
    }

    // Get nested power source input info
    CFDictionaryRef power_telemetry = get_dict_prop(properties, "PowerTelemetryData");
    if (power_telemetry) {
        info->source_voltage = get_long_prop(power_telemetry, "SystemVoltageIn");
        info->source_amperage = get_long_prop(power_telemetry, "SystemCurrentIn");
    }

    // --- End of data population ---

    CFRelease(properties); // Clean up the properties dictionary
    return 0; // Success
}

*/
import "C"
import (
	"fmt"
)

// GetBatteryInfo queries IOKit for all available power and battery telemetry
// and returns it in a structured format.
func GetBatteryInfo() (*BatteryInfo, error) {
	var c_info C.c_battery_info

	// Call the C function.
	ret := C.get_all_battery_info(&c_info)
	if ret != 0 {
		return nil, fmt.Errorf("IOKit query failed with C error code: %d", ret)
	}

	// The C call was successful, now we translate the C struct into our public Go struct.
	// This is where we also perform unit conversions (e.g., mV -> V).
	info := &BatteryInfo{
		IsCharging:   c_info.is_charging != 0,
		IsConnected:  c_info.is_connected != 0,
		FullyCharged: c_info.is_fully_charged != 0,

		Health: Health{
			CycleCount: int(c_info.cycle_count),
		},
		Capacity: Capacity{
			DesignCapacity:  int(c_info.design_capacity),
			MaxCapacity:     int(c_info.max_capacity),
			NominalCapacity: int(c_info.nominal_capacity),
		},
		Charge: Charge{
			CurrentCapacity: int(c_info.current_capacity),
			TimeToEmpty:     int(c_info.time_to_empty),
			TimeToFull:      int(c_info.time_to_full),
		},
		Temperature: Temperature{
			// IOKit reports temperature in hundredths of a degree Celsius.
			Battery: float64(c_info.temperature) / 100.0,
		},
		Power: Power{
			// IOKit reports voltage in mV and amperage in mA.
			Voltage:  float64(c_info.voltage) / 1000.0,
			Amperage: float64(c_info.amperage) / 1000.0,
		},
		Hardware: Hardware{
			SerialNumber: C.GoString(&c_info.serial_number[0]),
			DeviceName:   C.GoString(&c_info.device_name[0]),
		},
		Adapter: Adapter{
			Watts:       int(c_info.adapter_watts),
			Voltage:     float64(c_info.adapter_voltage) / 1000.0,
			Amperage:    float64(c_info.adapter_amperage) / 1000.0,
			Description: C.GoString(&c_info.adapter_description[0]),
		},
		PowerSourceInput: PowerSourceInput{
			Voltage:  float64(c_info.source_voltage) / 1000.0,
			Amperage: float64(c_info.source_amperage) / 1000.0,
		},
	}

	// We'll leave IndividualCellVoltages for a future iteration, as it requires parsing an array.
	// For now, it will be a nil slice.

	return info, nil
}

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
