package main

import (
	"github.com/pmurley/mida/log"
	"github.com/pmurley/mida/task"
	t "github.com/pmurley/mida/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"time"
)

// Fetch is the top level function of stage 1 of the MIDA pipeline and is responsible
// for getting the raw tasks (from any source) and placing them into the raw task channel.
func Fetch(rtc chan<- t.Task, cmd *cobra.Command, args []string) {
	switch cmd.Name() {
	case "file":
		rawTasks, err := task.ReadTasksFromFile(viper.GetString("task-file"))
		if err != nil {
			log.Error(err)
		}

		// Duplicate each task (or not) according to the "repeat" parameter
		var expandedTasks []t.Task
		for _, rt := range rawTasks {
			if rt.Repeat != nil {
				for i := 0; i < *rt.Repeat; i++ {
					expandedTasks = append(expandedTasks, rt)
				}
			} else {
				expandedTasks = append(expandedTasks, rt)
			}
		}

		// If option is enabled, shuffle our tasks so we execute them in random order
		if viper.GetBool("shuffle") {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(expandedTasks),
				func(i, j int) { expandedTasks[i], expandedTasks[j] = expandedTasks[j], expandedTasks[i] })
		}

		// Place tasks into the channel to stage 2 of the pipeline
		for _, rt := range expandedTasks {
			rtc <- rt
		}
	}

	// Close the task channel after we have dumped all tasks into it
	close(rtc)
}
