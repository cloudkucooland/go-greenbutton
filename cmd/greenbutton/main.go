package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cloudkucooland/go-greenbutton"
	"github.com/urfave/cli/v3"
)

type interval struct {
	Import float64 // kWh
	Export float64 // kWh
	Start  float64
}

func main() {
	cmd := &cli.Command{
		Name:  "greenbutton",
		Usage: "SMT Solar Plan Simulator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "plans",
				Value: "plans.json",
				Usage: "Path to plans JSON",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			plans, err := greenbutton.LoadPlans(cmd.String("plans"))
			if err != nil {
				return fmt.Errorf("load plans: %w", err)
			}

			monthly := make(map[string]map[time.Time]*interval)

			for _, filename := range cmd.Args().Slice() {
				f, err := os.Open(filename)
				if err != nil {
					return err
				}
				defer f.Close()

				handler := greenbutton.Parse
				if filepath.Ext(filename) == ".csv" {
					handler = greenbutton.ParseCSV
				}

				err = handler(f, func(r greenbutton.EnergyReading) {
					month := r.Start.Format("2006-01")

					if monthly[month] == nil {
						monthly[month] = make(map[time.Time]*interval)
					}

					if monthly[month][r.Start] == nil {
						monthly[month][r.Start] = &interval{}
					}

					kwh := r.Value / 1000.0 // Wh -> kWh

					if r.IsExport {
						monthly[month][r.Start].Export += kwh
					} else {
						monthly[month][r.Start].Import += kwh
					}
				})

				if err != nil {
					return err
				}
			}

			months := make([]string, 0, len(monthly))
			for m := range monthly {
				months = append(months, m)
			}
			sort.Strings(months)

			for _, p := range plans {
				fmt.Printf("\n--- Plan: %s ---\n", p.Name)
				fmt.Printf("%-8s | %-10s | %-10s | %-10s\n",
					"Month", "Import", "Export", "Cost ($)",
				)

				var makesTotalCents float64

				var monthlyBI billInterval
				for _, month := range months {
					var monthImport, monthExport float64
					var monthlyCents float64
					monthlyBI.EnergyChargeCents = 0
					monthlyBI.TDUChargeCents = 0
					monthlyBI.SolarCreditCents = 0

					for k, iv := range monthly[month] {
						// keep physics correct
						monthImport += iv.Import
						monthExport += iv.Export

						i := calculateInterval(p, iv.Import, iv.Export, k)
						monthlyBI.EnergyChargeCents += i.EnergyChargeCents
						monthlyBI.TDUChargeCents += i.TDUChargeCents
						monthlyBI.SolarCreditCents += i.SolarCreditCents
					}

					// monthly policy adjustments (caps, credits, etc.)
					monthlyCents = applyMonthlyRules(p, monthlyBI)

					makesTotalCents += monthlyCents

					fmt.Printf("%-8s | %10.1f | %10.1f | $%8.2f\n",
						month,
						monthImport,
						monthExport,
						monthlyCents/100,
					)
				}

				fmt.Printf("Total Projected Year Cost: $%.2f\n", makesTotalCents/100)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func applyMonthlyRules(p greenbutton.Plan, monthly billInterval) float64 {
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

type billInterval struct {
	EnergyChargeCents float64 // Gross Import * Energy Rate
	TDUChargeCents    float64 // Gross Import * TDU Rate
	SolarCreditCents  float64 // Gross Export * Buyback Rate
}

func calculateInterval(p greenbutton.Plan, imp, exp float64, t time.Time) billInterval {
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
	if p.Export.Model == greenbutton.ExportWholesale {
		exportRate = guessWholesaleFromHistorical(t)
	}

	return billInterval{
		EnergyChargeCents: billedImp * importRate,
		TDUChargeCents:    billedImp * tduRate,
		SolarCreditCents:  billedExp * exportRate,
	}
}
