package main

import "github.com/spf13/cobra"

func newDailyCommand() *cobra.Command {
	m := &cobra.Command{
		Use:   "daily",
		Short: "Daily Report",
		Args:  cobra.MinimumNArgs(0),
		Run:   runDailyCommandFunc,
	}

	return m
}

func runDailyCommandFunc(cmd *cobra.Command, args []string) {

}
