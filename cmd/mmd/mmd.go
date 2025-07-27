package main

import (
	"context"
	"fmt"
	"io"
	"os"

	mermaid_go "github.com/dreampuf/mermaid.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cmd := &cli.Command{
		Name:                  "mmd",
		Usage:                 "Render mermaid diagrams on the command line",
		Version:               fmt.Sprintf("%s - %s@%s", version, commit, date),
		EnableShellCompletion: true,
		HideHelpCommand:       true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "input",
				Value:     "",
				Aliases:   []string{"i"},
				Usage:     "Path to file that contains the mermaid diagram (if unset, read from stdin)",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:        "output",
				Value:       "",
				Aliases:     []string{"o"},
				DefaultText: "diagram.svg / diagram.png",
				Usage:       "Path of the output file",
			},
			&cli.StringFlag{
				Name:    "format",
				Value:   "svg",
				Aliases: []string{"f"},
				Usage:   "Output format. Possible values: svg, png",
			},
			&cli.FloatFlag{
				Name:    "scale",
				Value:   1.0,
				Aliases: []string{"s"},
				Usage:   "For PNG output only, scaling factor of the rendered image",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
				Usage: "Set log level: error, warn, info, debug, trace",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			err := Run(cmd)
			if err != nil {
				log.Error().Err(err).Msg("")
				os.Exit(1)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}

func Run(cmd *cli.Command) error {
	level, err := zerolog.ParseLevel(cmd.String("log-level"))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)

	re, _ := mermaid_go.NewRenderEngine(context.TODO())
	defer re.Cancel()

	file := cmd.String("input")

	var content, outputContent []byte
	if file == "" {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		content, err = os.ReadFile(file)
		if err != nil {
			return err
		}
	}

	output := cmd.String("output")
	switch cmd.String("format") {

	case "svg":
		log.Info().Msg("Rendering SVG")
		renderedContent, err := re.Render(string(content))
		outputContent = []byte(renderedContent)
		if err != nil {
			return err
		}
		if output == "" {
			output = "diagram.svg"
		}
	case "png":
		log.Info().Float64("scale", cmd.Float64("scale")).Msg("Rendering PNG")
		outputContent, _, err = re.RenderAsScaledPng(string(content), cmd.Float64("scale"))

		if err != nil {
			return err
		}
		if output == "" {
			output = "diagram.png"
		}
	default:
		return fmt.Errorf("unknown output format")
	}

	err = os.WriteFile(output, outputContent, 0644)
	if err != nil {
		return err
	}
	log.Info().Str("file", output).Msg("Content written")
	return nil
}
