package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInstanceKillsItselfDueToTimezoneBug(t *testing.T) {
	// Simulate the problematic scenario
	layoutNoTZ := "2006-01-02 15:04:05"

	// P1 (Old instance) starts at 10:00 CST
	p1Start := time.Date(2025, 4, 5, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	p1Formatted := p1Start.Format(layoutNoTZ) // "2025-04-05 10:00:00" (loses +08:00!)
	fmt.Printf("p1 starttime Formatted: %v\n", p1Formatted)

	// Network interruption happens during P2 startup
	// P2 (New instance) starts at 11:00 CST
	p2Start := time.Date(2025, 4, 5, 11, 0, 0, 0, time.FixedZone("CST", 8*3600))
	p2Formatted := p2Start.Format(layoutNoTZ) // "2025-04-05 11:00:00"
	fmt.Printf("p2 starttime Formatted: %v\n", p2Formatted)

	// Simulate name service restart and reconnection
	// P2 now discovers P1 from reconnected name service
	p1Parsed, err := time.Parse(layoutNoTZ, p1Formatted)
	assert.NoError(t, err)
	fmt.Printf("p2 parsed p1's starttime: %v\n", p1Parsed)

	// CRITICAL BUG: P1's time parsed as UTC!
	// "10:00" -> 10:00 UTC -> 18:00 CST
	expectedP1AsUTC := time.Date(2025, 4, 5, 10, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedP1AsUTC.Unix(), p1Parsed.Unix())

	// P2 compares times (in CST context):
	// P2 actual: 11:00 CST
	// P1 parsed: 18:00 CST (misinterpreted as UTC)
	// 11:00 < 18:00 -> P2 thinks P1 is NEWER!

	// Business logic: if other instance is newer, I should exit
	// P2 incorrectly concludes it should die
	shouldExit := p2Start.Before(p1Parsed) // 11:00 CST before 18:00 CST? TRUE!
	assert.True(t, shouldExit, "New instance incorrectly thinks it should exit!")

	// This causes the self-destruct behavior!
	t.Logf("BUG: P2 (started %s) thinks P1 (%s parsed as %s) is newer!",
		p2Start.Format("15:04 CST"),
		p1Start.Format("15:04 CST"),
		p1Parsed.In(time.FixedZone("CST", 8*3600)).Format("15:04 CST"))
}
