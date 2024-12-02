package models

import "github.com/google/uuid"

type DayOfWeek string

const (
	Monday    DayOfWeek = "monday"
	Tuesday   DayOfWeek = "tuesday"
	Wednesday DayOfWeek = "wednesday"
	Thursday  DayOfWeek = "thursday"
	Friday    DayOfWeek = "friday"
	Saturday  DayOfWeek = "saturday"
	Sunday    DayOfWeek = "sunday"
)

type Schedule struct {
	ScheduleID uuid.UUID
	DayOfWeek  DayOfWeek
	ItemID     uuid.UUID
	ChannelID  string
}
