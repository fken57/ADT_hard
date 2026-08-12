package training

import (
	"testing"
	"time"
)

func TestAcceptedDuringSessionUsesHalfOpenTimeRange(t *testing.T) {
	startedAt := time.Unix(1_000, 0).UTC()
	session := Session{StartedAt: startedAt, DurationSeconds: 4_500}
	tests := []struct {
		name       string
		submission Submission
		want       bool
	}{
		{name: "before start", submission: Submission{EpochSecond: 999, Result: "AC"}, want: false},
		{name: "at start", submission: Submission{EpochSecond: 1_000, Result: "AC"}, want: true},
		{name: "wrong result", submission: Submission{EpochSecond: 1_001, Result: "WA"}, want: false},
		{name: "before deadline", submission: Submission{EpochSecond: 5_499, Result: "AC"}, want: true},
		{name: "at deadline", submission: Submission{EpochSecond: 5_500, Result: "AC"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := AcceptedDuringSession(session, test.submission); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestAcceptedDuringSessionStopsAtAbortTime(t *testing.T) {
	startedAt := time.Unix(1_000, 0).UTC()
	abortedAt := time.Unix(1_100, 0).UTC()
	session := Session{StartedAt: startedAt, DurationSeconds: 4_500, Status: StatusAborted, EndedAt: &abortedAt}
	if !AcceptedDuringSession(session, Submission{EpochSecond: 1_099, Result: "AC"}) {
		t.Fatal("AC before abort should count")
	}
	if AcceptedDuringSession(session, Submission{EpochSecond: 1_100, Result: "AC"}) {
		t.Fatal("AC at or after abort must not count")
	}
}
