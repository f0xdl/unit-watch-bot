package domain

type Owner struct {
	UserId  int64    `json:"user_id"`
	Devices []string `json:"devices"` //UUIDs of devices
}
