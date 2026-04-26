package parser

import "sync/atomic"

// Counters tracks parser-level observability counters as required by §6.3 FP-011
// (counters: dropped, malformed, redacted, latency, rotation, reopen).
//
// All counter access is via atomic operations; safe for concurrent use.
type Counters struct {
	Malformed uint64 // §14.1 P-003 / P-007: lines that failed JSON / schema validation
	Redacted  uint64 // detector-side redaction events (rare; redaction is mainly logger-side)
	Detected  uint64 // detector classified the event as a non-zero risk_type
}

// IncMalformed increments the malformed counter atomically.
func (c *Counters) IncMalformed() { atomic.AddUint64(&c.Malformed, 1) }

// IncDetected increments the detected counter atomically.
func (c *Counters) IncDetected() { atomic.AddUint64(&c.Detected, 1) }

// IncRedacted increments the redacted counter atomically.
func (c *Counters) IncRedacted() { atomic.AddUint64(&c.Redacted, 1) }

// Snapshot returns a copy of the current counter values.
func (c *Counters) Snapshot() Counters {
	return Counters{
		Malformed: atomic.LoadUint64(&c.Malformed),
		Redacted:  atomic.LoadUint64(&c.Redacted),
		Detected:  atomic.LoadUint64(&c.Detected),
	}
}
