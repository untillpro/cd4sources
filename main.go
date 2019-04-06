package main

import (
	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

func main() {

	var cmdPull = &cobra.Command{
		Use:   "pull <main-repo> <repo2[=<repo2-fork]>",
		Short: "Pull sources from given repos",
		Args:  cobra.MinimumNArgs(1),
		Run:   runCmdPull,
	}
	cmdPull.PersistentFlags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
	cmdPull.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", ".", "Working directory")
	cmdPull.PersistentFlags().Int32VarP(&timeoutSec, "timeout", "t", 10, "Timeout")
	cmdPull.MarkPersistentFlagRequired("output")

	var rootCmd = &cobra.Command{Use: "directcd"}

	rootCmd.AddCommand(cmdPull)
	rootCmd.PersistentFlags().BoolVarP(&gc.IsVerbose, "verbose", "v", false, "Verbose output")
	rootCmd.Execute()

}
