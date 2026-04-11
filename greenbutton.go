package greenbutton

import (
	"encoding/xml"
	// "fmt"
	"io"
	"strings"
)

// Global state for a single Parse run
type parseState struct {
	meta   map[string]ReadingType     // Keyed by entry ID/link
	buffer map[string][]EnergyReading // Pending readings waiting for metadata
}

func Parse(r io.Reader, callback func(EnergyReading)) error {
	decoder := xml.NewDecoder(r)
	state := &parseState{
		meta:   make(map[string]ReadingType),
		buffer: make(map[string][]EnergyReading),
	}
	var activeTypeID string // Track the type for the current meter section

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if se, ok := token.(xml.StartElement); ok && se.Name.Local == "entry" {
			var e Entry
			if err := decoder.DecodeElement(&e, &se); err != nil {
				return err
			}

			// Capture the linkage from MeterReading entries to ReadingTypes
			for _, l := range e.Links {
				if l.Rel == "related" && strings.Contains(l.Href, "ReadingType") {
					activeTypeID = extractID(l.Href)
				}
			}

			// 1. If this is a ReadingType definition, store it
			if e.Content.ReadingType.FlowDirection != 0 {
				for _, l := range e.Links {
					if l.Rel == "self" {
						id := extractID(l.Href)
						state.meta[id] = e.Content.ReadingType
						state.flushBuffer(id, callback)
					}
				}
			}

			// 2. If this is data, use the activeTypeID we found earlier
			if len(e.Content.IntervalBlock.IntervalReading) > 0 {
				state.processBlock(activeTypeID, e.Content.IntervalBlock, callback)
			}
		}
	}
	return nil
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Entry struct {
	ID      string `xml:"id"`
	Title   string `xml:"title"`
	Links   []Link `xml:"link"` // Added this to capture all <link> tags
	Content struct {
		IntervalBlock struct {
			IntervalReading []struct {
				Value      int64 `xml:"value"`
				TimePeriod struct {
					Start int64 `xml:"start"`
				} `xml:"timePeriod"`
			} `xml:"IntervalReading"`
		} `xml:"IntervalBlock"`
		ReadingType ReadingType `xml:"ReadingType"`
	} `xml:"content"`
}

type ReadingType struct {
	FlowDirection        int `xml:"flowDirection"`        // 1=Import, 19=Export
	PowerOfTenMultiplier int `xml:"powerOfTenMultiplier"` // Scale factor
}
