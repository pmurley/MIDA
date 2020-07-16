package main

import (
	b "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/fetch"
	"github.com/pmurley/mida/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"time"
)

// stage1 is the top level function of stage 1 of the MIDA pipeline and is responsible
// for getting the raw tasks (from any source) and placing them into the raw task channel.
func stage1(rtc chan<- b.RawTask, cmd *cobra.Command, args []string) {
	switch cmd.Name() {
	case "file":
		rawTasks, err := fetch.FromFile(viper.GetString("task-file"))
		if err != nil {
			log.Log.Error(err)
			close(rtc)
			return
		}

		// If option is enabled, shuffle our tasks so we execute them in random order
		if viper.GetBool("shuffle") {
			// If we are shuffling tasks, we will need to read them all in to memory
			var taskSlice []b.RawTask
			for rt := range rawTasks {
				taskSlice = append(taskSlice, rt)
			}

			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(taskSlice),
				func(i, j int) { taskSlice[i], taskSlice[j] = taskSlice[j], taskSlice[i] })

			// Now that they are shuffled, put them in the channel
			for _, rt := range taskSlice {
				rtc <- rt
			}
		} else {
			// If we don'b need to shuffle, just pass the tasks through
			for rt := range rawTasks {
				rtc <- rt
			}
		}
	}

	// Close the task channel after we have dumped all tasks into it
	close(rtc)
}
