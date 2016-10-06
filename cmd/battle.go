package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/sean-jc/kraken/kraken"
)

// battleOpts aggregates the options that are battle to all actions
type battleOpts struct {
	output    string
	tentacles uint64
	cgroup    string
	args      string
}

func addBattleFlags(cmd *cobra.Command, opts *battleOpts) {
	cmd.Flags().StringVarP(&opts.output, "output", "o", "@", "output file name, \"@\" to use stdout (default)")
	cmd.Flags().Uint64VarP(&opts.tentacles, "tentacles", "t", 20, "number of hits (target processes) to spawn simultaneously")
	cmd.Flags().StringVarP(&opts.cgroup, "cgroup", "c", "", "name of cgroup to place child processes in")
	cmd.Flags().StringVarP(&opts.args, "args", "a", "", "arguments to pass to the target process (append to those from the target YAML)")
}

func newBattle(path string, opts *battleOpts, doBattle func(*kraken.Kraken, *kraken.Target) <-chan *kraken.Result) error {
	out, err := outputFile(opts.output)
	if err != nil {
		return fmt.Errorf("error opening %s: %s", opts.output, err)
	}
	defer out.Close()

	t, err := kraken.NewTarget(path, opts.args, opts.cgroup)
	if err != nil {
		return err
	}

	k, err := kraken.New(opts.tentacles)
	if err != nil {
		return err
	}

	res := doBattle(k, t)
	enc := kraken.NewEncoder(out)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-sig:
			k.Stop()
			return nil
		case r, ok := <-res:
			if !ok {
				return nil
			}
			if err = enc.Encode(r); err != nil {
				return err
			}
		}
	}
}
