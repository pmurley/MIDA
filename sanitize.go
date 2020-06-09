package main

import (
	"errors"
	"github.com/pmurley/mida/log"
	"github.com/pmurley/mida/task"
	t "github.com/pmurley/mida/types"
	"sync"
)

// SanitizeTasks (2) takes raw tasks from Fetch (1) stage and produces sanitized tasks for the Crawler stage (3)
func SanitizeTasks(rawTaskChan <-chan t.Task, sanitizedTaskChan chan<- t.TaskWrapper, pipelineWG *sync.WaitGroup) {
	for r := range rawTaskChan {
		st, err := sanitizeTask(r)
		if err != nil {
			log.Error(err)
			continue
		}
		pipelineWG.Add(1)

		sanitizedTaskChan <- st
	}

	// Wait until the pipeline is clear before we close the sanitized task channel,
	// which will cause MIDA to shutdown
	pipelineWG.Wait()
	close(sanitizedTaskChan)

	return
}

// sanitizeTask takes a raw tasks, checks it for validity, adds default values as needed,
// and creates a TaskWrapper object for it so it can be passed on for the site visit.
func sanitizeTask(rt t.Task) (t.TaskWrapper, error) {
	var tw t.TaskWrapper
	var err error

	tw.Task = task.AllocateNewTask()

	if rt.URL == nil || *rt.URL == "" {
		return tw, errors.New("invalid or missing URL for task")
	}

	*tw.Task.URL, err = ValidateURL(*rt.URL)
	if err != nil {
		return tw, err
	}

	log.Info(*tw.Task.URL)
	return tw, nil
}
