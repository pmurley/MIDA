package main

import (
	"github.com/pmurley/mida/log"
	t "github.com/pmurley/mida/types"
)

// Postprocess is stage 4 of the MIDA pipeline. It takes a RawResult produced by phase 3
// (which conducts the site visit) and conducts postprocessing to turn it into a FinalResult.
func Postprocess(rawResultChan <-chan t.RawResult, finalResultChan chan<- t.FinalResult) {
	for rawResult := range rawResultChan {
		fr, err := postprocessRawResult(rawResult)
		if err != nil {
			log.Error(err)
			continue
		}

		finalResultChan <- fr
	}

	// All PostProcessed results have been sent so close the channel
	close(finalResultChan)

	return
}

// postprocessRawResult is called by Postprocess (stage 4) to convert a RawResult
// to a FinalResult by applying postprocessing.
func postprocessRawResult(rr t.RawResult) (t.FinalResult, error) {
	log.Info(rr)
	var fr t.FinalResult
	return fr, nil
}
