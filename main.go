package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/kechako/goradio/audio"
	cli "github.com/urfave/cli/v2"
)

type ArgumentError string

func (err ArgumentError) Error() string {
	return string(err)
}

func run(ctx context.Context) (err error) {
	app := &cli.App{
		Name: "goradio",
		Commands: []*cli.Command{
			{
				Name:   "play",
				Usage:  "play radio",
				Action: playRadioCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "freq",
						Aliases:  []string{"f"},
						Usage:    "frequency to tune to (e.g. 93.0M, 90500K)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "device",
						Aliases:  []string{"d"},
						Usage:    "audio device name to play radio",
						Required: false,
					},
					&cli.IntFlag{
						Name:        "sample-rate",
						Aliases:     []string{"r"},
						Usage:       "audio sample rate",
						DefaultText: "default sample rate of audio device",
						Required:    false,
					},
					&cli.BoolFlag{
						Name:     "edge",
						Usage:    "enable lower edge tuning",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "dc",
						Usage:    "enable DC blocking filter",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "deemp",
						Usage:    "enable de-Emphasis filter",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "direct",
						Usage:    "enable direct sampling",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "offset",
						Usage:    "enable offset tuning",
						Required: false,
					},
				},
				OnUsageError: HandleUsageError,
			},
			{
				Name:         "record",
				Usage:        "record radio",
				Action:       recordRadioCommand,
				OnUsageError: HandleUsageError,
			},
			{
				Name:  "device",
				Usage: "show audio device information",
				Subcommands: []*cli.Command{
					{
						Name:         "list",
						Usage:        "list audio devices",
						Action:       deviceListCommand,
						OnUsageError: HandleUsageError,
					},
					{
						Name:         "show",
						Usage:        "show details of audio device",
						Action:       deviceShowCommand,
						OnUsageError: HandleUsageError,
					},
				},
				OnUsageError: HandleUsageError,
			},
		},
		Before: func(ctx *cli.Context) error {
			if err := audio.Initialize(); err != nil {
				return err
			}
			return nil
		},
		After: func(ctx *cli.Context) error {
			if err := audio.Terminate(); err != nil {
				return err
			}
			return nil
		},
		OnUsageError: HandleUsageError,
		ExitErrHandler: func(ctx *cli.Context, err error) {
			cli.HandleExitCoder(HandleError(ctx, err))
		},
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	return app.RunContext(ctx, os.Args)
}

func HandleUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	return cli.Exit(err, 2)
}

func HandleError(ctx *cli.Context, err error) error {
	if err == nil {
		return nil
	}

	var argErr ArgumentError
	if errors.As(err, &argErr) {
		return cli.Exit(argErr, 2)
	}

	var exitCoder cli.ExitCoder
	if errors.As(err, &exitCoder) {
		return exitCoder
	}

	return cli.Exit(err, 1)
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
