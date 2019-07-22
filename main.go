package main

import (
	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

func main() {

	var rootCmd = &cobra.Command{Use: "directcd"}
	rootCmd.PersistentFlags().BoolVarP(&gc.IsVerbose, "verbose", "v", false, "Verbose output")

	// cmdPull
	{
		var cmdPull = &cobra.Command{
			Use:   "pull --repo <main-repo> --replace <repo2=<repo-to-replace]> [args]",
			Short: "Periodically pull and build sources from given repos",
			Run:   runCmdPull,
		}

		cmdPull.PersistentFlags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
		cmdPull.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", ".", "Working directory")
		cmdPull.PersistentFlags().Int32VarP(&timeoutSec, "timeout", "t", 10, "Timeout")
		cmdPull.PersistentFlags().StringVarP(&mainRepo, "repo", "r", "", "Main repository")
		cmdPull.PersistentFlags().StringSliceVar(&argReplaced, "replace", []string{}, "Repositories to be replaced")
		cmdPull.MarkPersistentFlagRequired("output")
		cmdPull.MarkPersistentFlagRequired("repo")

		rootCmd.AddCommand(cmdPull)
	}

	rootCmd.Execute()

}

func generateCICommand() *cobra.Command {
	var cmdCD = &cobra.Command{
		Use:   "pull --repo <main-repo> --replace <repo2=<repo-to-replace]> [args]",
		Short: "Periodically pull and build sources from given repos",
	}

	cmdCD.PersistentFlags().StringVarP(&binaryName, "output", "o", "", "Output binary name")
	cmdCD.PersistentFlags().StringVarP(&workingDir, "working-dir", "w", ".", "Working directory")
	cmdCD.PersistentFlags().Int32VarP(&timeoutSec, "timeout", "t", 10, "Timeout")
	cmdCD.PersistentFlags().StringVarP(&mainRepo, "repo", "r", "", "Main repository")
	cmdCD.PersistentFlags().StringSliceVar(&argReplaced, "replace", []string{}, "Repositories to be replaced")
	cmdCD.MarkPersistentFlagRequired("output")
	cmdCD.MarkPersistentFlagRequired("repo")

	return cmdCD

}
