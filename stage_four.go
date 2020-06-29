package main

import (
	b "github.com/pmurley/mida/base"
	"github.com/pmurley/mida/log"
	"github.com/pmurley/mida/postprocess"
)

// stage4 is the postprocessing stage of the MIDA pipeline. It takes a RawResult produced by stage3
// (which conducts the site visit) and conducts postprocessing to turn it into a FinalResult.
func stage4(rawResultChan <-chan b.RawResult, finalResultChan chan<- b.FinalResult) {
	for rawResult := range rawResultChan {
		fr, err := postprocess.DevTools(&rawResult)
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
