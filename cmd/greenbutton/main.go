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

			var monthly greenbutton.MonthlyMap

			for _, filename := range cmd.Args().Slice() {
				f, err := os.Open(filename)
				if err != nil {
					return err
				}
				defer f.Close()

				// TODO this overwrites, rather than merges, fine for initial testing
				monthly, err = greenbutton.Loader(f, filepath.Ext(filename) == ".csv")
				if err != nil {
					return fmt.Errorf("load file: %w", err)
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
				for _, month := range months {
					mbi, monthlyCents, _ := monthly[month].Sum(context.Background(), p)
					makesTotalCents += monthlyCents

					fmt.Printf("%-8s | %10.1f | %10.1f | $%8.2f\n",
						month,
						mbi.Import,
						mbi.Export,
						monthlyCents/100,
					)
				}

				// need to normalize N months of data to yearly projection
				toYearCents := makesTotalCents / (float64(len(months)) / 12.0)
				fmt.Printf("Total Projected Year Cost: $%.2f\n", toYearCents/100)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
