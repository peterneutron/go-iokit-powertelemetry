// Package iokit provides direct access to macOS IOKit power and battery telemetry.
package power

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

	// Cell Voltages
    long cell_voltages[16]; // Assume max 16 cells, more than enough
    int  cell_voltage_count;

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

// Helper for parsing arrays.
static void get_long_array_prop(CFDictionaryRef dict, const char *key, long *out_array, int max_count, int *final_count) {
    *final_count = 0;
    CFStringRef key_ref = CFStringCreateWithCString(NULL, key, kCFStringEncodingUTF8);
    if (!key_ref) return;

    CFTypeRef value_ref = CFDictionaryGetValue(dict, key_ref);
    CFRelease(key_ref);

    if (value_ref != NULL && CFGetTypeID(value_ref) == CFArrayGetTypeID()) {
        CFArrayRef array_ref = (CFArrayRef)value_ref;
        CFIndex count = CFArrayGetCount(array_ref);
        if (count > max_count) {
            count = max_count; // Prevent buffer overflow
        }
        *final_count = (int)count;

        for (CFIndex i = 0; i < count; i++) {
            CFNumberRef num_ref = (CFNumberRef)CFArrayGetValueAtIndex(array_ref, i);
            if (num_ref != NULL && CFGetTypeID(num_ref) == CFNumberGetTypeID()) {
                CFNumberGetValue(num_ref, kCFNumberSInt64Type, &out_array[i]);
            } else {
                out_array[i] = 0; // Default value if type is wrong
            }
        }
    }
}

// The core C function to get all battery properties.
// Returns 0 on success, non-zero on error.
int get_all_battery_info(c_battery_info *info) {
    // Find the AppleSmartBattery service
    CFMutableDictionaryRef matching = IOServiceMatching("AppleSmartBattery");
    if (matching == NULL) return 1;

    io_iterator_t iterator;

	// IOServiceGetMatchingServices always consumes the 'matching' dictionary reference.
    if (IOServiceGetMatchingServices(kIOMainPortDefault, matching, &iterator) != KERN_SUCCESS) {
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

	// Get cell voltages from the nested BatteryData dictionary ---
    CFDictionaryRef battery_data = get_dict_prop(properties, "BatteryData");
    if (battery_data) {
        // We know CellVoltage is inside BatteryData
        get_long_array_prop(battery_data, "CellVoltage", info->cell_voltages, 16, &info->cell_voltage_count);
    }

    // --- End of data population ---

    CFRelease(properties); // Clean up the properties dictionary
    return 0; // Success
}

*/
import "C"
import (
	"fmt"
	"math"
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
		State: State{
			IsCharging:   c_info.is_charging != 0,
			IsConnected:  c_info.is_connected != 0,
			FullyCharged: c_info.is_fully_charged != 0,
		},
		Battery: Battery{
			SerialNumber:    C.GoString(&c_info.serial_number[0]),
			DeviceName:      C.GoString(&c_info.device_name[0]),
			CycleCount:      int(c_info.cycle_count),
			DesignCapacity:  int(c_info.design_capacity),
			MaxCapacity:     int(c_info.max_capacity),
			NominalCapacity: int(c_info.nominal_capacity),
			CurrentCapacity: int(c_info.current_capacity),
			TimeToEmpty:     int(c_info.time_to_empty),
			TimeToFull:      int(c_info.time_to_full),
			Temperature:     float64(c_info.temperature) / 100.0,
			Voltage:         float64(c_info.voltage) / 1000.0,
			Amperage:        float64(c_info.amperage) / 1000.0,
		},
		Adapter: Adapter{
			Description:   C.GoString(&c_info.adapter_description[0]),
			MaxWatts:      int(c_info.adapter_watts),
			MaxVoltage:    float64(c_info.adapter_voltage) / 1000.0,
			MaxAmperage:   float64(c_info.adapter_amperage) / 1000.0,
			InputVoltage:  float64(c_info.source_voltage) / 1000.0,
			InputAmperage: float64(c_info.source_amperage) / 1000.0,
		},
	}

	// Populate the individual cell voltages if they are available.
	if c_info.cell_voltage_count > 0 {
		// Create a Go slice of the exact correct size.
		info.Battery.IndividualCellVoltages = make([]int, c_info.cell_voltage_count)

		// Unsafe cast to a Go-accessible array pointer. This is a standard CGO pattern.
		c_voltages_ptr := &c_info.cell_voltages

		// Copy the values from the C array to our new Go slice.
		for i := 0; i < int(c_info.cell_voltage_count); i++ {
			info.Battery.IndividualCellVoltages[i] = int(c_voltages_ptr[i])
		}
	}

	// Calculate derived health metrics based on the collected data.
	calculateDerivedMetrics(info)
	return info, nil
}

// calculateDerivedMetrics populates the Calculations struct with health
// percentages and live power flow data in Watts.
func calculateDerivedMetrics(info *BatteryInfo) {
	// --- Health Percentage Calculations ---
	if info.Battery.DesignCapacity > 0 {
		designCapF := float64(info.Battery.DesignCapacity)

		healthByMax := (float64(info.Battery.MaxCapacity) / designCapF) * 100.0
		info.Calculations.HealthByMaxCapacity = int(math.Round(healthByMax))

		healthByNominal := (float64(info.Battery.NominalCapacity) / designCapF) * 100.0
		info.Calculations.HealthByNominalCapacity = int(math.Round(healthByNominal))

		var conditionModifier float64
		if len(info.Battery.IndividualCellVoltages) > 1 {
			minV, maxV := findMinMax(info.Battery.IndividualCellVoltages)
			drift := maxV - minV
			switch {
			case drift <= 5:
				conditionModifier = 2.5
			case drift <= 15:
				conditionModifier = 1.0
			case drift <= 30:
				conditionModifier = 0.0
			case drift <= 50:
				conditionModifier = -2.0
			default:
				conditionModifier = -10.0
			}
		}
		info.Calculations.ConditionAdjustedHealth = int(math.Round(healthByNominal + conditionModifier))
	}

	// --- Power Flow Calculations (Watts = Volts * Amps) ---

	// Helper function to truncate a float64 to two decimal places without rounding.
	truncate := func(f float64) float64 {
		return math.Trunc(f*100) / 100
	}

	// Power being drawn from the AC adapter.
	acPower := info.Adapter.InputVoltage * info.Adapter.InputAmperage
	info.Calculations.ACPower = truncate(acPower)

	// Power flowing into (+) or out of (-) the battery.
	batteryPower := info.Battery.Voltage * info.Battery.Amperage
	info.Calculations.BatteryPower = truncate(batteryPower)

	// The power consumed by the system (CPU, screen, etc.) is the combination of
	// power from the AC adapter and power from the battery.
	// If the battery is discharging, its power contribution is negative.
	systemPower := info.Calculations.ACPower - info.Calculations.BatteryPower
	info.Calculations.SystemPower = truncate(systemPower)
}

// Helper to find min/max in a slice
func findMinMax(a []int) (min int, max int) {
	if len(a) == 0 {
		return 0, 0
	}
	min = a[0]
	max = a[0]
	for _, value := range a {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

// BatteryInfo holds a comprehensive snapshot of all data points retrieved
// from the AppleSmartBattery service in IOKit.
type BatteryInfo struct {
	State        State
	Battery      Battery
	Adapter      Adapter
	Calculations Calculations
}

// State holds booleans describing the current charging status.
type State struct {
	IsCharging   bool
	IsConnected  bool
	FullyCharged bool
}

// Battery contains all data points directly related to the battery itself,
// from its hardware identifiers to its live electrical state.
type Battery struct {
	// Identity
	SerialNumber string
	DeviceName   string

	// Health & Capacity
	CycleCount      int
	DesignCapacity  int // in mAh
	MaxCapacity     int // in mAh
	NominalCapacity int // in mAh

	// Live Charge & Readings
	CurrentCapacity        int     // in mAh
	TimeToEmpty            int     // in minutes
	TimeToFull             int     // in minutes
	Temperature            float64 // in Celsius
	Voltage                float64 // in Volts
	Amperage               float64 // in Amps (negative when discharging)
	IndividualCellVoltages []int   // in mV
}

// Adapter holds information about the connected power source.
type Adapter struct {
	// Description is a system-provided string (e.g., "pd charger").
	Description string

	// MaxWatts is the negotiated power rating from the handshake (e.g., 96).
	MaxWatts int

	// MaxVoltage is the negotiated voltage from the handshake (e.g., 20.0V).
	MaxVoltage float64

	// MaxAmperage is the maximum current the adapter can provide at the
	// negotiated voltage (e.g., 4.8A).
	MaxAmperage float64

	// InputVoltage is the actual voltage being supplied by the adapter right now.
	InputVoltage float64

	// InputAmperage is the actual current being drawn by the system right now.
	InputAmperage float64
}

// Calculations contains derived, user-friendly metrics.
type Calculations struct {
	// Health percentages
	HealthByMaxCapacity     int
	HealthByNominalCapacity int
	ConditionAdjustedHealth int

	// Live power flow in Watts
	ACPower      float64 // Power being drawn from the AC adapter.
	BatteryPower float64 // Power flowing into(+) or out of(-) the battery.
	SystemPower  float64 // Power being consumed by the rest of the system.
}
