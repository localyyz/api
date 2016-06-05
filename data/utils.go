package data

import "time"

func GetTimeUTCPointer() *time.Time {
	t := time.Now().UTC()
	return &t
}
