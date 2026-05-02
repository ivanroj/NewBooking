// Package service contains pure business logic independent of database drivers.
package service

import "time"

// CheckOverlap reports whether two time intervals [s1,e1) and [s2,e2) overlap.
// Adjacent intervals (e1==s2) are NOT considered overlapping.
func CheckOverlap(s1, e1, s2, e2 time.Time) bool {
	return s1.Before(e2) && s2.Before(e1)
}

// ValidateBookingWindow returns false when start >= end or start is in the past.
func ValidateBookingWindow(start, end time.Time, now time.Time) bool {
	return start.Before(end) && start.After(now)
}
