package main

import (
	"github.com/pmurley/mida/log"
	"github.com/pmurley/mida/monitor"
	t "github.com/pmurley/mida/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"sync"
)

// InitPipeline is the main MIDA pipeline, used whenever MIDA uses a browser to visit websites.
// It consists of five main stages: Task Fetch, Task Sanitize, Site Visit, Postprocess, and Results Storage.
func InitPipeline(cmd *cobra.Command, args []string) {
	rawTaskChan := make(chan t.Task)              // channel connecting stages 1 and 2
	sanitizedTaskChan := make(chan t.TaskWrapper) // channel connecting stages 2 and 3
	rawResultChan := make(chan t.RawResult)       // channel connecting stages 3 and 4
	finalResultChan := make(chan t.FinalResult)   // channel connection stages 4 and 5
	retryChan := make(chan t.TaskWrapper)         // Channel connecting stage 5 and stage 3 for retrying failed tasks
	monitorChan := make(chan t.TaskSummary)

	var crawlerWG sync.WaitGroup  // Tracks active crawler workers
	var storageWG sync.WaitGroup  // Tracks active storage workers
	var pipelineWG sync.WaitGroup // Tracks tasks currently in pipeline

	// Start goroutine that runs the Prometheus monitoring HTTP server
	if viper.GetBool("monitor") {
		go monitor.RunPrometheusClient(monitorChan, viper.GetInt("prom-port"))
	}

	// Start goroutine(s) that handles crawl results storage
	numStorers := viper.GetInt("storers")
	storageWG.Add(numStorers)
	for i := 0; i < viper.GetInt("storers"); i++ {
		go Backend(finalResultChan, monitorChan, retryChan, &storageWG, &pipelineWG)
	}

	// Start goroutine that handles crawl results sanitization
	go Postprocess(rawResultChan, finalResultChan)

	// Start crawler(s) which take sanitized tasks as arguments
	numCrawlers := viper.GetInt("crawlers")
	crawlerWG.Add(numCrawlers)
	for i := 0; i < numCrawlers; i++ {
		go Crawler(sanitizedTaskChan, rawResultChan, retryChan, &crawlerWG)
	}

	// Start goroutine which sanitizes input tasks
	go SanitizeTasks(rawTaskChan, sanitizedTaskChan, &pipelineWG)

	// Start the goroutine responsible for getting our tasks
	go Fetch(rawTaskChan, cmd, args)

	// Wait for all of our crawlers to finish, and then allow them to exit
	crawlerWG.Wait()
	close(rawResultChan)

	// Wait for all of our storers to exit. We do not need to close the channel
	// going to storers -- the channel close will ripple through the pipeline
	storageWG.Wait()

	// Cleanup any remaining temporary files before we exit
	err := os.RemoveAll(TempDir)
	if err != nil {
		log.Error(err)
	}

	return
}
