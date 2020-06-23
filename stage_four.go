package main

import (
	b "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/log"
)

// stage4 is the postprocessing stage of the MIDA pipeline. It takes a RawResult produced by stage3
// (which conducts the site visit) and conducts postprocessing to turn it into a FinalResult.
func stage4(rawResultChan <-chan b.RawResult, finalResultChan chan<- b.FinalResult) {
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

// postprocessRawResult is called by stage4 (stage 4) to convert a RawResult
// to a FinalResult by applying postprocessing.
func postprocessRawResult(rr b.RawResult) (b.FinalResult, error) {
	var fr b.FinalResult
	return fr, nil
}
