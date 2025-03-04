package utils

import "time"

type DateTime struct{}

func NewDateTime() *DateTime {
	return &DateTime{}
}

func (DateTime) Now() time.Time {
	return time.Now()
}
