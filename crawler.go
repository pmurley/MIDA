package main

import (
	"github.com/pmurley/mida/log"
	t "github.com/pmurley/mida/types"
	"sync"
)

func Crawler(taskWrapperChan <-chan t.TaskWrapper, rawResultChan chan<- t.RawResult, retryChan <-chan t.TaskWrapper, crawlerWG *sync.WaitGroup) {

	for tw := range taskWrapperChan {
		rawResult, err := executeSiteVisit(tw)
		if err != nil {
			log.WithURL(*tw.Task.URL).Error(err)
			continue
		}

		rawResultChan <- rawResult
	}

	crawlerWG.Done()
}

func executeSiteVisit(tw t.TaskWrapper) (t.RawResult, error) {
	var rr t.RawResult
	log.Info(tw)
	return rr, nil
}
