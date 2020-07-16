package main

import (
	t "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/log"
	"github.com/pmurley/mida/storage"
	"sync"
)

func stage5(finalResultChan <-chan t.FinalResult, monitoringChan chan<- t.TaskSummary,
	retryChan chan<- t.TaskWrapper, storageWG *sync.WaitGroup, pipelineWG *sync.WaitGroup) {

	for fr := range finalResultChan {
		err := storage.Store(fr)
		if err != nil {
			log.Log.Error(err)
		}

		pipelineWG.Done()
	}

	storageWG.Done()
}
