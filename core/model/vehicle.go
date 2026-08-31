package model

import "time"

type Vehicle struct {
	ID          ID
	FleetID     ID
	Type        string
	PlateNumber string
	DeviceTime  time.Time
	Attributes  JSON
}
