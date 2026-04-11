package greenbutton

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"
)

// Smart Meter Texas format

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

		// TODO we need to collapse the samples to 15-minute buckets since sometimes
		// export is 09:15:01 and import is 09:15:02
		start, err := time.Parse("01/02/2006 15:04", v[1]+" "+v[3])
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(v[5], 64)
		if err != nil {
			continue
		}

		callback(EnergyReading{
			Start:    start,
			Value:    val * 1000, // Convert kWh to Wh for consistency
			IsExport: v[7] == "Surplus Generation",
		})
	}
	return nil
}
