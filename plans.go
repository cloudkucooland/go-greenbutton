package greenbutton

import (
	"encoding/json"
	"os"
)

type Plan struct {
	Name      string  `json:"Name"`
	Base      float64 `json:"Base"`      // cents
	ExportKwh float64 `json:"ExportKwh"` // cents per kwh (rebate)
	ImportKwh float64 `json:"ImportKwh"` // cents per kwh (price)
	TDUBase   float64 `json:"TDUBase"`   // cents
	TDUKwh    float64 `json:"TDUKwh"`    // cents per kwh
	MaxKwhNet int     `json:"MaxKwhNet"` // -1 for no limit
}

func LoadPlans(path string) ([]Plan, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var plans []Plan
	return plans, json.Unmarshal(b, &plans)
}
