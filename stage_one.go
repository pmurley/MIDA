package main

import (
	"github.com/pmurley/mida/amqp"
	b "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/fetch"
	"github.com/pmurley/mida/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// stage1 is the top level function of stage 1 of the MIDA pipeline and is responsible
// for getting the raw tasks (from any source) and placing them into the raw task channel.
func stage1(rtc chan<- *b.RawTask, cmd *cobra.Command, args []string) {
	switch cmd.Name() {
	case "file":
		rawTasks, err := fetch.FromFile(viper.GetString("task-file"), viper.GetBool("shuffle"))
		if err != nil {
			log.Log.Error(err)
			close(rtc)
			return
		}
		for rt := range rawTasks {
			rtc <- rt
		}

	case "go":
		// Generate our task set from command line options, decompress it,
		// and load our tasks into the pipeline.
		cts, err := BuildCompressedTaskSet(cmd, args)
		if err != nil {
			log.Log.Error(err)
			close(rtc)
			return
		}

		rawTasks := b.ExpandCompressedTaskSet(*cts)
		for _, rt := range rawTasks {
			rtc <- &rt
		}

	case "client":
		var params = amqp.ConnParams{
			User: viper.GetString("amqp-user"),
			Pass: viper.GetString("amqp-pass"),
			Uri:  viper.GetString("amqp-uri"),
		}

		taskAMQPConn, taskDeliveryChan, err := amqp.NewAMQPTasksConsumer(params, viper.GetString("amqp-task-queue"))
		if err != nil {
			log.Log.Fatal(err)
		}
		defer taskAMQPConn.Shutdown()

		broadcastAMQPConn, broadcastAMQPDeliveryChan, err := amqp.NewAMQPBroadcastConsumer(params, amqp.DefaultBroadcastExchange)
		if err != nil {
			log.Log.Fatal(err)
		}
		defer broadcastAMQPConn.Shutdown()

		log.Log.Infof("Successfully connected to AMQP Queue: \"%s\"", viper.GetString("amqp-task-queue"))

		// Remain as a client to the AMQP server until a broadcast is received which
		// causes us to exit
		breakFlag := false
		for {
			select {
			case broadcastMsg := <-broadcastAMQPDeliveryChan:
				log.Log.Warnf("BROADCAST RECEIVED: [ %s ]", string(broadcastMsg.Body))
				if string(broadcastMsg.Body) == "quit" {
					breakFlag = true
				}
			default:
			}
			select {
			case broadcastMsg := <-broadcastAMQPDeliveryChan:
				log.Log.Warnf("BROADCAST RECEIVED: [ %s ]", string(broadcastMsg.Body))
				if string(broadcastMsg.Body) == "quit" {
					breakFlag = true
				}
			case amqpMsg := <-taskDeliveryChan:
				rawTask, err := amqp.DecodeAMQPMessageToRawTask(amqpMsg)
				if err != nil {
					log.Log.Error(err)
				}
				rtc <- &rawTask
			}
			if breakFlag {
				break
			}
		}
	}

	// Close the task channel after we have dumped all tasks into it
	close(rtc)
}
