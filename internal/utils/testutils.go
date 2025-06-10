package utils

import (
	"testing"
	"time"
)

// AssertEventually retries cond until it returns true or times out.
// Fails the test if condition not met in time.
func AssertEventually(t *testing.T, d time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("condition not met within %v", d)
}
