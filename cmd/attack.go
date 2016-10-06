package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/sean-jc/kraken/kraken"
)

type attackOpts struct {
	rate     uint64
	duration uint
	battle   battleOpts
}

func attackCmd() *cobra.Command {
	var opts attackOpts

	cmd := &cobra.Command{
		Use:  "attack [OPTIONS] TARGET",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doBattle := func(k *kraken.Kraken, t *kraken.Target) <-chan *kraken.Result {
				return k.Attack(t, opts.rate, time.Duration(opts.duration)*time.Millisecond)
			}
			return newBattle(args[0], &opts.battle, doBattle)
		},
	}

	addBattleFlags(cmd, &opts.battle)

	cmd.Flags().UintVarP(&opts.duration, "duration", "d", 20*1000, "duration of the attack, in milliseconds")
	cmd.Flags().Uint64VarP(&opts.rate, "rate", "r", 20, "rate of attack, in hits per second")

	return cmd
}
