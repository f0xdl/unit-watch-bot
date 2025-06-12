package domain

import "time"

type Device struct {
	UUID      string       `json:"uuid"`
	Label     string       `json:"label"`
	Active    bool         `json:"active"`
	Status    DeviceStatus `json:"status"`
	OwnerId   int64        `json:"owner_id"`
	LastSeen  time.Time    `json:"online_at"`
	PointId   int          `json:"point_id"`
	ExpiresAt time.Time    `json:"expires_at"`
}

func (d *Device) Online() bool {
	return time.Now().Sub(d.LastSeen).Minutes() < 5
}
