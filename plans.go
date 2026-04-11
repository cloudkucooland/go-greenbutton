package greenbutton

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Plan struct {
	Name string

	Charges Charges

	Export struct {
		Model ExportModel

		FixedRate float64 // for ExportFixed
		// wholesale handled separately later
	}

	Netting NettingRules
	Credits CreditBank
	Limits  ExportLimits

	TOU     *TimeOfUse
	Battery *BatteryRules
}

type Charges struct {
	BaseCents    float64 // REP base
	TDUBaseCents float64

	ImportCentsPerKWh float64
	TDUCentsPerKWh    float64
}

type ExportModel string

const (
	ExportNone      ExportModel = "none"
	ExportNetting   ExportModel = "netting"   // 1:1
	ExportFixed     ExportModel = "fixed"     // flat rate
	ExportWholesale ExportModel = "wholesale" // RTW index
)

type NettingRules struct {
	Mode string // "instant", "monthly", "billing_cycle"

	// If true: export can only offset import (no net positive payout)
	NoNetExport bool

	// If true: export credit limited to import volume
	CapToImport bool
}

type CreditBank struct {
	Enabled bool

	// expires after N months (0 = never)
	ExpirationMonths int

	// whether credits can reduce fixed charges
	AppliesToBase bool
}

type ExportLimits struct {
	MaxKWhPerMonth   float64
	MaxCreditDollars float64
}

type TimeOfUse struct {
	Enabled bool
	Periods []TOUPeriod
}

type TOUPeriod struct {
	Name       string    `json:"name"`
	ImportRate float64   `json:"import_rate"`
	ExportRate float64   `json:"export_rate"`
	Start      TimeOfDay `json:"start"` // "16:00" in JSON
	End        TimeOfDay `json:"end"`   // "21:00" in JSON
}

type BatteryRules struct {
	Required            bool
	ExportOnlyFromSolar bool // Tesla-style restrictions
}

func LoadPlans(path string) ([]Plan, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var plans []Plan
	return plans, json.Unmarshal(b, &plans)
}

type TimeOfDay int // Minutes since midnight

func (t *TimeOfDay) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	parsed, err := time.Parse("15:04", s)
	if err != nil {
		return fmt.Errorf("invalid time format %s: use HH:MM", s)
	}

	*t = TimeOfDay(parsed.Hour()*60 + parsed.Minute())
	return nil
}

func (tp TOUPeriod) ActiveAt(t time.Time) bool {
	mins := t.Hour()*60 + t.Minute()
	start := int(tp.Start)
	end := int(tp.End)

	if start < end {
		return mins >= start && mins < end
	}
	// Handle overnight periods (e.g., 22:00 to 06:00)
	return mins >= start || mins < end
}
