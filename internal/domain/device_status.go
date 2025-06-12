package domain

type DeviceStatus int

const (
	StatusUnknown DeviceStatus = iota
	StatusInactive
	StatusNormal
	StatusError
)

func ParseDeviceStatus(status int) DeviceStatus {
	if 0 >= status || status > 3 {
		return StatusUnknown
	}
	return DeviceStatus(status)
}

func (ds DeviceStatus) String() string {
	switch ds {
	case StatusInactive:
		return "Inactive"
	case StatusNormal:
		return "Normal"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}
