package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sean-jc/kraken/kraken"
)

type pummelOpts struct {
	hits   uint64
	battle battleOpts
}

func pummelCmd() *cobra.Command {
	var opts pummelOpts

	cmd := &cobra.Command{
		Use:  "pummel",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.hits == 0 {
				return fmt.Errorf("hits must be >0")
			}

			doBattle := func(k *kraken.Kraken, t *kraken.Target) <-chan *kraken.Result {
				return k.Pummel(t, opts.hits)
			}
			return newBattle(args[0], &opts.battle, doBattle)
		},
	}

	addBattleFlags(cmd, &opts.battle)

	cmd.Flags().Uint64VarP(&opts.hits, "hits", "s", 20, "total number of hits on the target")

	return cmd
}
