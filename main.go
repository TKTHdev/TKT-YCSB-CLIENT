package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "ycsb-client",
		Usage: "YCSB benchmark client — sends GET/SET over TCP to a key-value server",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "addr",
				Usage:    "Server address(es) (e.g. localhost:7000). Can be specified multiple times for multiple servers.",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "workers",
				Usage: "Number of concurrent worker goroutines",
				Value: 1,
			},
			&cli.StringFlag{
				Name:  "workload",
				Usage: "YCSB workload: ycsb-a (50% writes), ycsb-b (5% writes), ycsb-c (read-only)",
				Value: "ycsb-a",
			},
			&cli.IntFlag{
				Name:  "keys",
				Usage: "Number of distinct keys",
				Value: 6,
			},
			&cli.StringFlag{
				Name:  "csv",
				Usage: "Path to CSV file for result output (appends; creates with header row if new)",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug logging",
			},
		},
		Action: func(c *cli.Context) error {
			cfg := Config{
				Addrs:    c.StringSlice("addr"),
				Workers:  c.Int("workers"),
				NumKeys:  c.Int("keys"),
				Workload: ParseWorkload(c.String("workload")),
				CSVPath:  c.String("csv"),
				Debug:    c.Bool("debug"),
			}
			Run(cfg)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
