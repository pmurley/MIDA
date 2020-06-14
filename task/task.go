package task

import (
	"encoding/json"
	"errors"
	"github.com/pmurley/mida/log"
	t "github.com/pmurley/mida/types"
	"io/ioutil"
)

// ReadTasksFromFile is a wrapper function that reads single tasks, full task sets,
// or compressed task sets from file.
func ReadTasksFromFile(filename string) ([]t.RawTask, error) {
	tasks := make(t.TaskSet, 0)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return tasks, errors.New("failed to read task file: " + filename)
	}

	tasks, err = ReadTasksFromBytes(data)
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

// WriteTaskSliceToFile takes a RawTask slice and writes it out as a JSON file to a given filename.
func WriteTaskSliceToFile(tasks []t.RawTask, filename string) error {
	taskBytes, err := WriteTaskSliceToBytes(tasks)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, taskBytes, 0644)
	return err
}

// WriteCompressedTaskSetToFile takes a CompressedTaskSet and writes a JSON representation
// of it out to a file
func WriteCompressedTaskSetToFile(tasks t.CompressedTaskSet, filename string) error {
	taskBytes, err := WriteCompressedTaskSetToBytes(tasks)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, taskBytes, 0644)
	return err
}

// ExpandCompressedTaskSet takes a CompressedTaskSet object and converts it into a slice
// of regular Tasks.
func ExpandCompressedTaskSet(ts t.CompressedTaskSet) []t.RawTask {
	var rawTasks []t.RawTask

	repeats := 1
	if ts.Repeat != nil && *ts.Repeat > 0 {
		repeats = *ts.Repeat
	}
	for i := 0; i < repeats; i += 1 {
		for _, singleUrl := range *ts.URL {
			var url = singleUrl
			newTask := t.RawTask{
				URL:         &url,
				Browser:     ts.Browser,
				Completion:  ts.Completion,
				Data:        ts.Data,
				Output:      ts.Output,
				MaxAttempts: ts.MaxAttempts,
			}
			rawTasks = append(rawTasks, newTask)
		}
	}
	return rawTasks
}

// ReadTasksFromBytes reads in tasks from a byte array. It will read them whether they
// are formatted as individual tasks or as a CompressedTaskSet.
func ReadTasksFromBytes(data []byte) ([]t.RawTask, error) {
	tasks := make(t.TaskSet, 0)
	err := json.Unmarshal(data, &tasks)
	if err == nil {
		log.Debugf("Parsed TaskSet (%d tasks) from file", len(tasks))
		return tasks, nil
	}

	var singleTask t.RawTask
	err = json.Unmarshal(data, &singleTask)
	if err == nil {
		log.Debug("Parsed single RawTask from file")
		return append(tasks, singleTask), nil
	}

	compressedTaskSet := t.CompressedTaskSet{}
	err = json.Unmarshal(data, &compressedTaskSet)
	if err != nil {
		return tasks, errors.New("failed to unmarshal tasks: [ " + err.Error() + " ]")
	}

	if compressedTaskSet.URL == nil || len(*compressedTaskSet.URL) == 0 {
		return tasks, errors.New("no URLs given in task set")
	}
	tasks = ExpandCompressedTaskSet(compressedTaskSet)

	log.Debugf("Parsed CompressedTaskSet (%d tasks) from file", len(tasks))
	return tasks, nil

}

// WriteTaskSliceToBytes takes a slice of tasks and converts it to corresponding JSON bytes to transfer somewhere.
func WriteTaskSliceToBytes(tasks []t.RawTask) ([]byte, error) {
	taskBytes, err := json.Marshal(tasks)
	if err != nil {
		return nil, err
	}

	return taskBytes, nil
}

// WriteCompressedTaskSetToBytes takes a CompressedTaskSet and converts it to corresponding JSON bytes to transfer somewhere.
func WriteCompressedTaskSetToBytes(tasks t.CompressedTaskSet) ([]byte, error) {
	taskBytes, err := json.Marshal(tasks)
	if err != nil {
		return nil, err
	}

	return taskBytes, nil
}

// AllocateNewTask allocates a new RawTask struct, initializing everything to zero values
func AllocateNewTask() *t.RawTask {
	var task = new(t.RawTask)
	task.URL = new(string)
	task.MaxAttempts = new(int)

	task.Browser = AllocateNewBrowserSettings()
	task.Completion = AllocateNewCompletionSettings()
	task.Data = AllocateNewDataSettings()
	task.Output = AllocateNewOutputSettings()

	return task
}

// AllocateNewBrowserSettings allocates a new BrowserSettings struct, initializing everything to zero values
func AllocateNewBrowserSettings() *t.BrowserSettings {
	var bs = new(t.BrowserSettings)
	bs.BrowserBinary = new(string)
	bs.AddBrowserFlags = new([]string)
	bs.RemoveBrowserFlags = new([]string)
	bs.SetBrowserFlags = new([]string)
	bs.Extensions = new([]string)
	bs.UserDataDirectory = new(string)

	return bs
}

// AllocateNewCompletionSettings allocates a new CompletionSettings struct, initializing everything to zero values
func AllocateNewCompletionSettings() *t.CompletionSettings {
	var cs = new(t.CompletionSettings)
	cs.TimeAfterLoad = new(int)
	cs.Timeout = new(int)
	cs.CompletionCondition = new(t.CompletionCondition)

	return cs
}

// AllocateNewDataSettings allocates a new DataSettings struct, initializing everything to zero values
func AllocateNewDataSettings() *t.DataSettings {
	var ds = new(t.DataSettings)
	ds.AllResources = new(bool)
	ds.ResourceMetadata = new(bool)

	return ds
}

// AllocateNewOutputSettings allocates a new OutputSettings struct, initializing everything to zero values
func AllocateNewOutputSettings() *t.OutputSettings {
	var ops = new(t.OutputSettings)
	ops.LocalOut = new(t.LocalOutputSettings)
	ops.LocalOut.DS = AllocateNewDataSettings()
	ops.LocalOut.Path = new(string)

	ops.SftpOut = new(t.SftpOutputSettings)
	ops.SftpOut.DS = AllocateNewDataSettings()
	ops.SftpOut.Host = new(string)
	ops.SftpOut.Path = new(string)
	ops.SftpOut.Port = new(int)
	ops.SftpOut.UserName = new(string)
	ops.SftpOut.PrivateKeyFile = new(string)

	return ops
}
