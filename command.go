package main

import (
	"github.com/pmurley/mida/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getRootCommand returns the root cobra command which will be executed based on arguments passed to the program
func getRootCommand() *cobra.Command {
	var cmdRoot = &cobra.Command{Use: "mida"}

	var (
		numCrawlers int
		numStorers  int
		monitor     bool
		promPort    int
		logLevel    int
	)

	cmdRoot.PersistentFlags().IntVarP(&numCrawlers, "crawlers", "c", viper.GetInt("crawlers"),
		"Number of parallel browser instances to use for crawling")
	cmdRoot.PersistentFlags().IntVarP(&numStorers, "storers", "s", viper.GetInt("storers"),
		"Number of parallel goroutines working to store task results")
	cmdRoot.PersistentFlags().BoolVarP(&monitor, "monitor", "m", false,
		"Enable monitoring via Prometheus by hosting a HTTP server")
	cmdRoot.PersistentFlags().IntVarP(&promPort, "prom-port", "p", viper.GetInt("prom-port"),
		"Port used for hosting metrics for a Prometheus server")
	cmdRoot.PersistentFlags().IntVarP(&logLevel, "log-level", "l", viper.GetInt("log-level"),
		"Log Level for MIDA (0=Error, 1=Warn, 2=Info, 3=Debug)")

	cmdRoot.AddCommand(getFileCommand())

	return cmdRoot
}

// getFileCommand returns the command for the "mida file" version of the program
func getFileCommand() *cobra.Command {
	var cmdFile = &cobra.Command{
		Use:   "file",
		Short: "Read and execute tasks from file",
		Long: `MIDA reads and executes tasks from a pre-created task
file, exiting when all tasks in the file are completed.`,
		Run: func(cmd *cobra.Command, args []string) {
			ll, err := cmd.Flags().GetInt("log-level")
			if err != nil {
				log.Fatal(err)
			}
			err = log.ConfigureLogging(ll)
			if err != nil {
				log.Fatal(err)
			}

			InitPipeline(cmd, args)
		},
	}

	var (
		taskFile string
		shuffle  bool
	)

	cmdFile.Flags().StringVarP(&taskFile, "task-file", "f", viper.GetString("task-file"),
		"RawTask file to process")
	cmdFile.Flags().BoolVarP(&shuffle, "shuffle", "", DefaultShuffle,
		"Randomize processing order for tasks")
	err := viper.BindPFlag("task-file", cmdFile.Flags().Lookup("task-file"))
	if err != nil {
		log.Fatal(err)
	}

	// Enable some autocomplete features of Cobra
	_ = cmdFile.MarkFlagFilename("task-file")

	return cmdFile
}
