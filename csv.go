package greenbutton

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"
)

// Smart Meter Texas CSV format
func ParseCSV(r io.Reader, callback func(EnergyReading)) error {
	reader := csv.NewReader(r)
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, v := range lines {
		if v[0] == "ESIID" {
			continue
		} // Skip header

		// all sample data shows an even nn:[00|15|30|45]:00 time
		start, err := time.Parse("01/02/2006 15:04", v[1]+" "+v[3])
		if err != nil {
			continue
		}

		// not currently used, for future work if we need more integration
		end, err := time.Parse("01/02/2006 15:04", v[1]+" "+v[4])
		if err != nil {
			continue
		}

		val, err := strconv.ParseFloat(v[5], 64)
		if err != nil {
			continue
		}

		callback(EnergyReading{
			Start:    start,
			End:      end,
			ValueKWh: val,
			IsExport: v[7] == "Surplus Generation",
		})
	}
	return nil
}
