package main

import (
	"github.com/pmurley/mida/amqp"
	b "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/teamnsrg/mida/storage"
)

// getRootCommand returns the root cobra command which will be executed based on arguments passed to the program
func getRootCommand() *cobra.Command {
	var err error
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

	err = viper.BindPFlags(cmdRoot.PersistentFlags())
	if err != nil {
		log.Log.Fatal("viper failed to bind pflags")
	}

	cmdRoot.AddCommand(getBuildCommand())
	cmdRoot.AddCommand(getFileCommand())
	cmdRoot.AddCommand(getLoadCommand())

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
				log.Log.Fatal(err)
			}
			err = log.ConfigureLogging(ll)
			if err != nil {
				log.Log.Fatal(err)
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
	cmdFile.Flags().BoolVarP(&shuffle, "shuffle", "", b.DefaultShuffle,
		"Randomize processing order for tasks")
	err := viper.BindPFlag("task-file", cmdFile.Flags().Lookup("task-file"))
	if err != nil {
		log.Log.Fatal(err)
	}

	// Enable some autocomplete features of Cobra
	_ = cmdFile.MarkFlagFilename("task-file")

	return cmdFile
}

func getLoadCommand() *cobra.Command {
	var cmdLoad = &cobra.Command{
		Use:   "load",
		Short: "Load tasks from file into queue",
		Long:  `Read tasks from a JSON-formatted file, parse them, and load them into the specified queue instance`,
		Args:  cobra.ExactArgs(1), // the filename containing tasks to read
		Run: func(cmd *cobra.Command, args []string) {
			ll, err := cmd.Flags().GetInt("log-level")
			if err != nil {
				log.Log.Fatal(err)
			}
			err = log.ConfigureLogging(ll)
			if err != nil {
				log.Log.Fatal(err)
			}

			tasks, err := b.ReadTasksFromFile(args[0])
			if err != nil {
				log.Log.Fatal(err)
			}

			var params = amqp.ConnParams{
				User: viper.GetString("amqpuser"),
				Pass: viper.GetString("amqppass"),
				Host: viper.GetString("amqphost"),
				Port: viper.GetInt("amqpport"),
				Tls:  viper.GetBool("tls"),
			}

			numTasksLoaded, err := amqp.LoadTasks(tasks, params, viper.GetString("queue"),
				uint8(viper.GetInt("priority")), viper.GetBool("shuffle"))
			if err != nil {
				log.Log.Fatal(err)
			}

			log.Log.Infof("Loaded %d tasks into queue \"%s\" with priority %d",
				numTasksLoaded, viper.GetString("queue"), viper.GetInt("priority"))
		},
	}

	var (
		taskFile string
		shuffle  bool
		queue    string
		priority uint8
	)

	cmdLoad.Flags().StringVarP(&taskFile, "task-file", "f", viper.GetString("task-file"),
		"Task file to process")
	cmdLoad.Flags().StringVarP(&queue, "queue", "q", amqp.DefaultQueue,
		"AMQP queue into which we will load tasks")
	cmdLoad.Flags().BoolVarP(&shuffle, "shuffle", "", b.DefaultShuffle,
		"Randomize loading order for tasks")
	cmdLoad.Flags().BoolVarP(&shuffle, "tls", "", amqp.DefaultTls,
		"Randomize loading order for tasks")
	cmdLoad.Flags().Uint8VarP(&priority, "priority", "p", amqp.DefaultPriority,
		"Priority of tasks we are loaded (AMQP: x-max-priority setting)")

	// Enable some autocomplete features of Cobra
	_ = cmdLoad.MarkFlagFilename("task-file")

	// We have to get a task file
	_ = cmdLoad.MarkFlagRequired("task-file")

	return cmdLoad
}

func getBuildCommand() *cobra.Command {
	// Variables storing options for the build command
	var (
		urlFile  string
		priority int

		// Browser settings
		browser            string
		userDataDir        string
		addBrowserFlags    []string
		removeBrowserFlags []string
		setBrowserFlags    []string
		extensions         []string

		// Completion settings
		completionCondition string
		timeout             int
		timeAfterLoad       int

		// Data Gathering settings
		resourceMetadata bool
		allResources     bool

		// Output settings
		resultsOutputPath string // Results from task path

		outputPath string // Task file path
		overwrite  bool

		// How many times a task should be repeated
		repeat int
	)

	var cmdBuild = &cobra.Command{
		Use:   "build",
		Short: "Build a MIDA Task File",
		Long:  `Create and save a task file using flags or CLI`,
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ll, err := cmd.Flags().GetInt("log-level")
			if err != nil {
				log.Log.Fatal(err)
			}
			err = log.ConfigureLogging(ll)
			if err != nil {
				log.Log.Fatal(err)
			}
			cts, err := BuildCompressedTaskSet(cmd, args)
			if err != nil {
				log.Log.Error(err)
				return
			}
			err = writeCompressedTaskSet(cts, cmd)
			if err != nil {
				log.Log.Error(err)
			}
		},
	}

	cmdBuild.Flags().StringVarP(&urlFile, "url-file", "f",
		"", "File containing URL to visit (1 per line)")
	cmdBuild.Flags().IntVarP(&priority, "priority", "", b.DefaultTaskPriority,
		"Task priority (when loaded into RabbitMQ")

	cmdBuild.Flags().StringVarP(&browser, "browser", "b",
		"", "Path to browser binary to use for this task")
	cmdBuild.Flags().StringVarP(&userDataDir, "user-data-dir", "d",
		"", "User Data Directory used for this task.")
	cmdBuild.Flags().StringSliceP("add-browser-flags", "", addBrowserFlags,
		"Flags to add to browser launch (comma-separated, no '--')")
	cmdBuild.Flags().StringSliceP("remove-browser-flags", "", removeBrowserFlags,
		"Flags to remove from browser launch (comma-separated, no '--')")
	cmdBuild.Flags().StringSliceP("set-browser-flags", "", setBrowserFlags,
		"Overrides default browser flags (comma-separated, no '--')")
	cmdBuild.Flags().StringSliceP("extensions", "e", extensions,
		"Full paths to browser extensions to use (comma-separated, no'--')")

	cmdBuild.Flags().StringVarP(&completionCondition, "completion", "y", string(b.DefaultCompletionCondition),
		"Completion condition for tasks (CompleteOnTimeoutOnly, CompleteOnLoadEvent, CompleteOnTimeoutAfterLoad")
	cmdBuild.Flags().IntVarP(&timeout, "timeout", "t", b.DefaultTimeout,
		"Timeout (in seconds) after which the browser will close and the task will complete")
	cmdBuild.Flags().IntVarP(&timeAfterLoad, "time-after-load", "", b.DefaultTimeAfterLoad,
		"Time after load event to remain on page (overridden by timeout if reached first)")

	cmdBuild.Flags().BoolVarP(&allResources, "all-resources", "", b.DefaultAllResources,
		"Gather and store all resources downloaded by browser")
	cmdBuild.Flags().BoolVarP(&resourceMetadata, "resource-metadata", "", b.DefaultResourceMetadata,
		"Gather and store metadata about all resources downloaded by browser")

	cmdBuild.Flags().StringVarP(&resultsOutputPath, "results-output-path", "r", storage.DefaultOutputPath,
		"Path (local or remote) to store results in. A new directory will be created inside this one for each task.")

	cmdBuild.Flags().StringVarP(&outputPath, "outfile", "o", viper.GetString("task-file"),
		"Path to write the newly-created JSON task file")
	cmdBuild.Flags().BoolVarP(&overwrite, "overwrite", "x", false,
		"Allow overwriting of an existing task file")

	cmdBuild.Flags().IntVarP(&repeat, "repeat", "", 1,
		"How many times to repeat a given task")

	_ = cmdBuild.MarkFlagRequired("url-file")
	_ = cmdBuild.MarkFlagFilename("url-file")

	return cmdBuild
}
