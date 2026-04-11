package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/cloudkucooland/go-greenbutton"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "greenbutton",
		Usage: "SMT Solar Plan Simulator",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "plans", Value: "plans.json", Usage: "Path to plans JSON"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			plans, err := greenbutton.LoadPlans(cmd.String("plans"))
			if err != nil {
				return fmt.Errorf("failed to load plans: %w", err)
			}

			// Aggregate by Month -> {Import, Export}
			type monthData struct{ Imp, Exp float64 }
			monthly := make(map[string]*monthData)

			for _, filename := range cmd.Args().Slice() {
				f, _ := os.Open(filename)
				defer f.Close()

				handler := greenbutton.Parse
				if filepath.Ext(filename) == ".csv" {
					handler = greenbutton.ParseCSV
				}

				handler(f, func(r greenbutton.EnergyReading) {
					key := r.Start.Format("2006-01")
					if monthly[key] == nil { monthly[key] = &monthData{} }
					if r.IsExport {
						monthly[key].Exp += r.Value / 1000.0 // to kWh
					} else {
						monthly[key].Imp += r.Value / 1000.0
					}
				})
			}

			// Run Simulation
			for _, p := range plans {
				fmt.Printf("\n--- Plan: %s ---\n", p.Name)
				fmt.Printf("%-8s | %-8s | %-8s | %-8s\n", "Month", "Import", "Export", "Cost ($)")
				
				var totalCost float64
				keys := make([]string, 0, len(monthly))
				for k := range monthly { keys = append(keys, k) }
				sort.Strings(keys)

				for _, k := range keys {
					m := monthly[k]
					// Your simulation logic
					cost := (p.Base + p.TDUBase)
					cost += (m.Imp * p.ImportKwh)
					cost += (m.Imp * p.TDUKwh)
					
					// Credit for export
					credit := m.Exp * p.ExportKwh
					if p.MaxKwhNet != -1 && (m.Exp - m.Imp) > float64(p.MaxKwhNet) {
						credit = float64(p.MaxKwhNet) * p.ExportKwh
					}
					
					finalMonth := (cost - credit) / 100.0
					totalCost += finalMonth
					fmt.Printf("%-8s | %6.1f kW | %6.1f kW | $%8.2f\n", k, m.Imp, m.Exp, finalMonth)
				}
				fmt.Printf("Total Projected Year Cost: $%.2f\n", totalCost)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
