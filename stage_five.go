package main

import (
	t "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/log"
	"sync"
)

func Backend(finalResultChan <-chan t.FinalResult, monitoringChan chan<- t.TaskSummary,
	retryChan chan<- t.TaskWrapper, storageWG *sync.WaitGroup, pipelineWG *sync.WaitGroup) {

	for fr := range finalResultChan {
		err := storeResults(fr)
		if err != nil {
			log.Error(err)
		}

		pipelineWG.Done()
	}

	storageWG.Done()
}

func storeResults(fr t.FinalResult) error {
	return nil
}
