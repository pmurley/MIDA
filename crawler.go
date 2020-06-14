package main

import (
	t "github.com/pmurley/mida/types"
	"sync"
)

func Crawler(taskWrapperChan <-chan t.TaskWrapper, rawResultChan chan<- t.RawResult, retryChan <-chan t.TaskWrapper, crawlerWG *sync.WaitGroup) {

	for tw := range taskWrapperChan {
		rawResult, err := executeSiteVisit(tw)
		if err != nil {
			continue
		}

		rawResultChan <- rawResult
	}

	crawlerWG.Done()
}

func executeSiteVisit(tw t.TaskWrapper) (t.RawResult, error) {
	var rr t.RawResult
	return rr, nil
}
