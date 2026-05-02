package service_test

import (
	"testing"
	"time"

	"github.com/example/coworking/internal/service"
)

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCheckOverlap_NoConflict(t *testing.T) {
	// [09:00-10:00) vs [10:00-11:00) — adjacent, no overlap
	if service.CheckOverlap(ts("2026-05-10T09:00:00Z"), ts("2026-05-10T10:00:00Z"),
		ts("2026-05-10T10:00:00Z"), ts("2026-05-10T11:00:00Z")) {
		t.Fatal("adjacent intervals must NOT overlap")
	}
}

func TestCheckOverlap_Conflict(t *testing.T) {
	// [09:00-11:00) vs [10:00-12:00) — overlap [10:00-11:00)
	if !service.CheckOverlap(ts("2026-05-10T09:00:00Z"), ts("2026-05-10T11:00:00Z"),
		ts("2026-05-10T10:00:00Z"), ts("2026-05-10T12:00:00Z")) {
		t.Fatal("overlapping intervals must conflict")
	}
}

func TestCheckOverlap_OneContainsAnother(t *testing.T) {
	// [09:00-18:00) contains [11:00-13:00)
	if !service.CheckOverlap(ts("2026-05-10T09:00:00Z"), ts("2026-05-10T18:00:00Z"),
		ts("2026-05-10T11:00:00Z"), ts("2026-05-10T13:00:00Z")) {
		t.Fatal("contained interval must conflict")
	}
}

func TestCheckOverlap_Identical(t *testing.T) {
	s := ts("2026-05-10T09:00:00Z")
	e := ts("2026-05-10T10:00:00Z")
	if !service.CheckOverlap(s, e, s, e) {
		t.Fatal("identical intervals must conflict")
	}
}

func TestValidateBookingWindow_Valid(t *testing.T) {
	now := ts("2026-05-10T08:00:00Z")
	if !service.ValidateBookingWindow(ts("2026-05-10T09:00:00Z"), ts("2026-05-10T11:00:00Z"), now) {
		t.Fatal("expected valid booking window")
	}
}

func TestValidateBookingWindow_StartAfterEnd(t *testing.T) {
	now := ts("2026-05-10T08:00:00Z")
	if service.ValidateBookingWindow(ts("2026-05-10T11:00:00Z"), ts("2026-05-10T09:00:00Z"), now) {
		t.Fatal("start > end must be invalid")
	}
}

func TestValidateBookingWindow_StartInPast(t *testing.T) {
	now := ts("2026-05-10T12:00:00Z")
	if service.ValidateBookingWindow(ts("2026-05-10T09:00:00Z"), ts("2026-05-10T11:00:00Z"), now) {
		t.Fatal("start in the past must be invalid")
	}
}
