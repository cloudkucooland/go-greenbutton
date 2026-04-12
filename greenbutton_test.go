package greenbutton

import (
	// "context"
	// "errors"
	"fmt"
	"os"
	// "strings"
	"testing"
	"time"
)

type testfiles struct {
	path       string
	shouldFail bool
}

var tf = []testfiles{
	{"tests/good.xml", false}, {"tests/big.xml", false}, {"tests/daily.xml", false},
}

func TestStaticParse(t *testing.T) {
	// func Parse(r io.Reader, callback func(EnergyReading)) error
	for _, tt := range tf {
		t.Run(tt.path, func(t *testing.T) {
			reader, err := os.Open(tt.path)
			if err != nil {
				t.Fatalf("data not found")
			}
			err = Parse(reader, func(enre EnergyReading) {
				prefix := " [IMPORT]"
				if enre.IsExport {
					prefix = " [EXPORT]"
				}
				// Format the time to something human-friendly (RFC822 is short and clean)
				fmt.Printf("%s %s: %.2f KWh\n", prefix, enre.Start.Format(time.RFC822), enre.ValueKWh)
			})

			if !tt.shouldFail && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.shouldFail && err == nil {
				t.Fatalf("should have failed")
			}
		})
	}
}

func TestInstantNetting(t *testing.T) {
	p := Plan{
		Name:    "Instant 1:1 Test",
		Charges: Charges{ImportCentsPerKWh: 20.0, TDUCentsPerKWh: 5.0},
		Netting: NettingRules{Mode: "instant"},
		Export: struct {
			Model     ExportModel
			FixedRate float64
		}{Model: ExportNetting, FixedRate: 20.0},
	}

	// Case: 2kWh Import, 3kWh Export in one interval
	// Result: Net 1kWh Export. Should only be charged TDU on the NET import (0).
	// Credit should be 1kWh * 20c = 20c.
	res := p.CalculateInterval(2.0, 3.0, time.Now())

	if res.EnergyChargeCents != 0 {
		t.Errorf("Expected 0 energy charge for net export, got %.2f", res.EnergyChargeCents)
	}
	if res.SolarCreditCents != 20.0 {
		t.Errorf("Expected 20c solar credit, got %.2f", res.SolarCreditCents)
	}
}

func TestNoNetExport(t *testing.T) {
	p := Plan{
		Name:    "NNE Test",
		Charges: Charges{BaseCents: 1000, ImportCentsPerKWh: 10.0},
		Netting: NettingRules{NoNetExport: true},
	}

	// Monthly summary: $50 in energy charges, $100 in solar credits.
	monthly := MonthlyBillInterval{
		EnergyChargeCents: 5000,
		SolarCreditCents:  10000,
		TDUChargeCents:    0,
	}

	total := p.ApplyMonthlyRules(monthly)

	// Expected: Energy (50) - Credit (100) = -50 -> Capped to 0.
	// Total = 0 + 1000 (Base) = 1000 cents ($10).
	if total != 1000 {
		t.Errorf("Expected 1000 cents (base only), got %.2f", total)
	}
}

func TestTOUActivity(t *testing.T) {
	// Saturday at 2:00 PM
	testTime := time.Date(2026, 4, 11, 14, 0, 0, 0, time.Local)

	period := TOUPeriod{
		Name:  "Weekend",
		Days:  []string{"Saturday", "Sunday"},
		Start: 0,    // 00:00
		End:   1439, // 23:59
	}

	if !period.ActiveAt(testTime) {
		t.Error("Failed to identify Saturday afternoon as a Weekend TOU period")
	}
}
