package domain

import "time"

type Device struct {
	UUID      string       `json:"uuid"`
	Label     string       `json:"label"`
	Active    bool         `json:"active"`
	Status    DeviceStatus `json:"status"`
	OwnerId   int64        `json:"owner_id"`
	LastSeen  time.Time    `json:"last_seen"`
	Point     string       `json:"point"`
	ExpiresAt time.Time    `json:"expires_at"`
	Online    bool         `json:"online"`
}
