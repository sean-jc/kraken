package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sean-jc/kraken/kraken"
)

type reportOpts struct {
	format string
	input  string
	output string
}

func reportCmd() *cobra.Command {
	var opts reportOpts

	cmd := &cobra.Command{
		Use:  "report [OPTIONS]",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return report(&opts)
		},
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "format of output [text, json, hist[buckets]]")
	cmd.Flags().StringVarP(&opts.input, "input", "i", "@", "input file names (comma separated), \"@\" to use stdin (default) ")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "@", "output file name, \"@\" to use stdout (default)")

	return cmd
}

// report validates the report arguments, sets up the required resources
// and writes the report
func report(opts *reportOpts) error {
	if len(opts.format) < 4 {
		return fmt.Errorf("Invalid format %s, options are [text, json, hist[buckets]]", opts.format)
	}

	in := strings.Split(opts.input, ",")
	srcs := make([]io.Reader, len(in))
	for i, n := range in {
		f, err := inputFile(n)
		if err != nil {
			return err
		}
		defer f.Close()
		srcs[i] = f
	}
	dec := kraken.NewDecoder(srcs...)

	out, err := outputFile(opts.output)
	if err != nil {
		return err
	}
	defer out.Close()

	var rep kraken.Reporter
	var report kraken.Report

	switch opts.format[:4] {
	case "text":
		var m kraken.Metrics
		rep, report = kraken.NewTextReporter(&m), &m
	case "json":
		var m kraken.Metrics
		rep, report = kraken.NewJSONReporter(&m), &m
	// case "plot":
	// 	var rs kraken.Results
	// 	rep, report = kraken.NewPlotReporter("kraken Plot", &rs), &rs
	case "hist":
		if len(opts.format) < 6 {
			return fmt.Errorf("Bad histogram format, expect at least two buckets, received: '%s'", opts.format[4:])
		}
		var hist kraken.Histogram
		if err := hist.Buckets.UnmarshalText([]byte(opts.format[4:])); err != nil {
			return err
		}
		rep, report = kraken.NewHistogramReporter(&hist), &hist
	default:
		return fmt.Errorf("Unknown format: %q", opts.format)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

decode:
	for {
		select {
		case <-sigch:
			break decode
		default:
			var r kraken.Result
			if err = dec.Decode(&r); err != nil {
				if err == io.EOF {
					break decode
				}
				return err
			}
			report.Add(&r)
		}
	}

	if c, ok := report.(kraken.Closer); ok {
		c.Close()
	}

	return rep.Report(out)
}
