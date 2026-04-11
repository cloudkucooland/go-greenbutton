package greenbutton

import (
	"strings"
	"time"
)

// EnergyReading is the clean output
type EnergyReading struct {
	Start time.Time
	// Duration time.Duration
	Value    float64
	IsExport bool
}

func (s *parseState) processBlock(typeID string, block IntervalBlock, callback func(EnergyReading)) {
	meta, ok := s.meta[typeID]
	if !ok {
		// Metadata hasn't arrived yet; buffer it
		s.buffer[typeID] = append(s.buffer[typeID], blockToReadings(block)...)
		return
	}

	for _, r := range blockToReadings(block) {
		callback(s.applyMeta(r, meta))
	}
}

func (s *parseState) flushBuffer(typeID string, callback func(EnergyReading)) {
	meta := s.meta[typeID]
	for _, r := range s.buffer[typeID] {
		callback(s.applyMeta(r, meta))
	}
	delete(s.buffer, typeID)
}

func (s *parseState) applyMeta(r EnergyReading, meta ReadingType) EnergyReading {
	// Apply scale (PowerOfTenMultiplier) and detect flow
	// Flow 19 = Export/Surplus, Flow 1 = Import/Consumption
	r.IsExport = (meta.FlowDirection == 19)

	// SMT often uses PowerOfTenMultiplier -3 (1234 -> 1.234 kWh)
	// or 0 (1234 -> 1234 Wh).
	// Let's assume the value is in Wh for consistency.
	return r
}

func blockToReadings(block IntervalBlock) []EnergyReading {
	var out []EnergyReading
	for _, ir := range block.IntervalReading {
		out = append(out, EnergyReading{
			Start: time.Unix(ir.TimePeriod.Start, 0),
			Value: float64(ir.Value),
		})
	}
	return out
}

func extractID(href string) string {
	parts := strings.Split(href, "/")
	return parts[len(parts)-1]
}

// Ensure these match your XML tags exactly
type IntervalBlock struct {
	IntervalReading []struct {
		Value      int64 `xml:"value"`
		TimePeriod struct {
			Start int64 `xml:"start"`
		} `xml:"timePeriod"`
	} `xml:"IntervalReading"`
}
