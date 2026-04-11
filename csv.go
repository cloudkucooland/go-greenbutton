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
		if v[0] == "ESIID" { continue } // Skip header
		
		// SMT CSV Date format: 04/01/2026 and 00:00
		start, _ := time.Parse("01/02/2006 15:04", v[1]+" "+v[3])
		val, _ := strconv.ParseFloat(v[5], 64)

		callback(EnergyReading{
			Start:    start,
			Value:    val * 1000, // Convert kWh to Wh for consistency
			IsExport: v[7] == "Surplus Generation",
		})
	}
	return nil
}
