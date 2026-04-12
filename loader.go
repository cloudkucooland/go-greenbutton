package greenbutton

import (
	"context"
	"io"
	"time"
)

type Interval struct {
	Import float64 // kWh
	Export float64 // kWh
	Start  float64
}

type IntervalMap map[time.Time]*Interval
type MonthlyMap map[string]IntervalMap

func Loader(f io.Reader, isCSV bool) (MonthlyMap, error) {
	monthly := make(MonthlyMap)

	handler := Parse
	if isCSV {
		handler = ParseCSV
	}

	err := handler(f, func(r EnergyReading) {
		month := r.Start.Format("2006-01")

		if monthly[month] == nil {
			monthly[month] = make(IntervalMap)
		}

		if monthly[month][r.Start] == nil {
			monthly[month][r.Start] = &Interval{}
		}

		if r.IsExport {
			monthly[month][r.Start].Export += r.ValueKWh
		} else {
			monthly[month][r.Start].Import += r.ValueKWh
		}
	})

	if err != nil {
		return nil, err
	}
	return monthly, nil
}

func (im IntervalMap) Sum(ctx context.Context, p Plan) (MonthlyBillInterval, float64, error) {
	bi := MonthlyBillInterval{0, 0, 0, 0, 0}
	var monthlyCents float64

	var skipper int
	for k, iv := range im {
		if skipper%96 == 0 {
			select {
			case <-ctx.Done():
				return MonthlyBillInterval{}, 0, ctx.Err()
			default:
			}
		}
		skipper++

		// keep physics correct
		bi.Import += iv.Import
		bi.Export += iv.Export

		i := p.CalculateInterval(iv.Import, iv.Export, k)
		bi.EnergyChargeCents += i.EnergyChargeCents
		bi.TDUChargeCents += i.TDUChargeCents
		bi.SolarCreditCents += i.SolarCreditCents
	}

	// monthly policy adjustments (caps, credits, etc.)
	monthlyCents = p.ApplyMonthlyRules(bi)
	return bi, monthlyCents, nil
}

func (p Plan) ApplyMonthlyRules(monthly MonthlyBillInterval) float64 {
	energyCharges := monthly.EnergyChargeCents
	solarCredits := monthly.SolarCreditCents
	tduCharges := monthly.TDUChargeCents

	// CapToImport: Some plans only credit you up to the amount of kWh you bought
	// NoNetExport: You can't have a negative bill (Chariot)

	netEnergy := energyCharges - solarCredits

	if p.Netting.NoNetExport && netEnergy < 0 {
		netEnergy = 0
	}

	// Total = Net Energy + Non-bypassable TDU + Base Fees
	// note: for a partial month of data we are still adding full monthly base
	// for our purposes that is fine, no need to prorate unless we just want
	// to overachive later
	return netEnergy + tduCharges + p.Charges.BaseCents + p.Charges.TDUBaseCents
}

// historicalRTW represents the estimated solar-weighted wholesale
// buyback rates in cents per kWh for the ERCOT North Hub.
var historicalRTW = map[int]float64{
	1:  3.50, // January: Winter peaks
	2:  4.10, // February: Volatility risk
	3:  0.56, // March: Low load, high renewables (Verified by Bill)
	4:  1.20, // April: Shoulder month
	5:  2.10, // May: Early heat
	6:  2.98, // June: 2024 Actual
	7:  2.16, // July: 2024 Actual
	8:  2.43, // August: Summer weighted (Verified by Bill)
	9:  2.34, // September: 2024 Actual
	10: 1.50, // October: Shoulder month
	11: 1.30, // November: Low load
	12: 2.20, // December: Winter demand
}

func guessWholesaleFromHistorical(t time.Time) float64 {
	historical, ok := historicalRTW[int(t.Month())]
	if !ok {
		return 1.0 // sane default?
	}
	return historical // * 0.8 - round down?
}

type BillInterval struct {
	EnergyChargeCents float64 // Gross Import * Energy Rate
	TDUChargeCents    float64 // Gross Import * TDU Rate
	SolarCreditCents  float64 // Gross Export * Buyback Rate
}

type MonthlyBillInterval struct {
	EnergyChargeCents float64 // Gross Import * Energy Rate
	TDUChargeCents    float64 // Gross Import * TDU Rate
	SolarCreditCents  float64 // Gross Export * Buyback Rate
	Import            float64
	Export            float64
}

func (p Plan) CalculateInterval(imp, exp float64, t time.Time) BillInterval {
	importRate := p.Charges.ImportCentsPerKWh
	tduRate := p.Charges.TDUCentsPerKWh
	exportRate := p.Export.FixedRate

	// Apply TOU overrides if active for this interval
	if p.TOU != nil && p.TOU.Enabled {
		for _, tou := range p.TOU.Periods {
			if tou.ActiveAt(t) {
				importRate = tou.ImportRate
				exportRate = tou.ExportRate
				break
			}
		}
	}

	var billedImp, billedExp float64
	if p.Netting.Mode == "instant" {
		net := imp - exp
		if net > 0 {
			billedImp = net
			billedExp = 0
		} else {
			billedImp = 0
			billedExp = -net
			// For instant 1:1, the credit for the netted surplus is usually the retail rate
			exportRate = importRate
		}
	} else {
		billedImp = imp
		billedExp = exp
	}

	// Handle Wholesale model
	if p.Export.Model == ExportWholesale {
		exportRate = guessWholesaleFromHistorical(t)
	}

	return BillInterval{
		EnergyChargeCents: billedImp * importRate,
		TDUChargeCents:    billedImp * tduRate,
		SolarCreditCents:  billedExp * exportRate,
	}
}
