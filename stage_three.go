package main

import (
	t "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/browser"
	"sync"
)

func stage3(taskWrapperChan <-chan t.TaskWrapper, rawResultChan chan<- t.RawResult, retryChan <-chan t.TaskWrapper, crawlerWG *sync.WaitGroup) {

	for tw := range taskWrapperChan {
		rawResult, err := browser.VisitPageDevtoolsProtocol(tw)
		if err != nil {
			break
		}

		rawResultChan <- *rawResult
	}

	crawlerWG.Done()
}
