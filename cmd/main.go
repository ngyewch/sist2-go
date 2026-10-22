package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	version string

	inputFileArg = &cli.StringArg{
		Name:      "input-file",
		UsageText: "(input-file)",
	}
	outputFileArg = &cli.StringArg{
		Name:      "output-file",
		UsageText: "(output-file)",
	}

	app = &cli.Command{
		Name:    "sist2-go",
		Usage:   "sist2-go",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:   "convert",
				Action: doConvert,
				Arguments: []cli.Argument{
					inputFileArg,
					outputFileArg,
				},
			},
		},
	}
)

func main() {
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		slog.Error("error",
			slog.Any("err", err),
		)
		os.Exit(1)
	}
}
