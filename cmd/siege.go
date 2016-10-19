package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/sean-jc/kraken/kraken"
)

type siegeOpts struct {
	duration uint
	battle   battleOpts
}

func siegeCmd() *cobra.Command {
	var opts siegeOpts

	cmd := &cobra.Command{
		Use:  "siege [OPTIONS] TARGET",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doBattle := func(k *kraken.Kraken, t *kraken.Target) <-chan *kraken.Result {
				return k.Siege(t, time.Duration(opts.duration)*time.Second)
			}
			return newBattle(args[0], &opts.battle, doBattle)
		},
	}

	addBattleFlags(cmd, &opts.battle)

	cmd.Flags().UintVarP(&opts.duration, "duration", "d", 20, "duration of the attack, in seconds")

	return cmd
}
