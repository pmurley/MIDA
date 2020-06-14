package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/pmurley/mida/log"
	"github.com/pmurley/mida/task"
	t "github.com/pmurley/mida/types"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// SanitizeTasks (2) takes raw tasks from Fetch (1) stage and produces sanitized tasks for the Crawler stage (3)
func SanitizeTasks(rawTaskChan <-chan t.RawTask, sanitizedTaskChan chan<- t.TaskWrapper, pipelineWG *sync.WaitGroup) {
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
func sanitizeTask(rt t.RawTask) (t.TaskWrapper, error) {
	var tw t.TaskWrapper
	var err error

	// Each task gets its own UUID
	tw.UUID = uuid.New()

	if rt.URL == nil || *rt.URL == "" {
		return t.TaskWrapper{}, errors.New("missing or empty URL for task")
	}

	tw.SanitizedTask.URL, err = ValidateURL(*rt.URL)
	if err != nil {
		return t.TaskWrapper{}, err
	}

	tw.SanitizedTask.BrowserBinaryPath, err = getBrowserBinaryPath(rt)
	if err != nil {
		return t.TaskWrapper{}, err
	}

	tw.SanitizedTask.BrowserFlags, err = getBrowserFlags(rt)
	if err != nil {
		return t.TaskWrapper{}, err
	}

	tw.SanitizedTask.UserDataDirectory, err = getUserDataDirectory(rt, tw.UUID)
	if err != nil {
		return t.TaskWrapper{}, err
	}

	tw.SanitizedTask.CS, err = sanitizeCompletionSettings(rt.Completion)
	if err != nil {
		return t.TaskWrapper{}, err
	}

	tw.SanitizedTask.DS, err = sanitizeDataSettings(rt.Data, DefaultDataSettings)

	return tw, nil
}

// getBrowserBinaryPath uses input from the task to sanitize and set the full path to the browser
// binary we will use for this crawl. If an invalid path is provided, it returns an error. If no
// path is provided, it attempts to select a default.
// Order of preference:
//   1. Chromium
//   2. Chrome
func getBrowserBinaryPath(rt t.RawTask) (string, error) {

	if rt.Browser == nil || rt.Browser.BrowserBinary == nil || *rt.Browser.BrowserBinary == "" {
		if runtime.GOOS == "darwin" {
			if _, err := os.Stat(DefaultOSXChromiumPath); err == nil {
				return DefaultOSXChromiumPath, nil
			} else if _, err := os.Stat(DefaultOSXChromePath); err == nil {
				return DefaultOSXChromePath, nil
			} else {
				return "", errors.New("no browser binary provided and could not find a default")
			}
		} else if runtime.GOOS == "linux" {
			if _, err := os.Stat(DefaultLinuxChromiumPath); err == nil {
				return DefaultLinuxChromiumPath, nil
			} else if _, err := os.Stat(DefaultLinuxChromePath); err == nil {
				return DefaultLinuxChromePath, nil
			} else {
				return "", errors.New("no browser binary provided and could not find a default")
			}
		} else if runtime.GOOS == "windows" {
			if _, err := os.Stat(DefaultWindowsChromiumPath); err == nil {
				return DefaultWindowsChromiumPath, nil
			} else if _, err := os.Stat(DefaultWindowsChromePath); err == nil {
				return DefaultWindowsChromePath, nil
			} else {
				return "", errors.New("no browser binary provided and could not find a default")
			}
		} else {
			return "", errors.New("this operating system is not supported by MIDA (MIDA supports Windows, Linux, Mac OS)")
		}
	} else {
		_, err := os.Stat(*rt.Browser.BrowserBinary)
		if err != nil {
			return *rt.Browser.BrowserBinary, nil
		}

		// We were given a browser string that is not a path to a file that exists
		// We offer some shortcuts for popular browsers
		if strings.ToLower(*rt.Browser.BrowserBinary) == "chrome" {
			if _, err := os.Stat(DefaultOSXChromePath); err == nil && runtime.GOOS == "darwin" {
				return DefaultOSXChromePath, nil
			} else if _, err := os.Stat(DefaultLinuxChromePath); err == nil && runtime.GOOS == "linux" {
				return DefaultLinuxChromePath, nil
			} else if _, err := os.Stat(DefaultWindowsChromePath); err == nil && runtime.GOOS == "windows" {
				return DefaultWindowsChromePath, nil
			} else {
				return "", errors.New("could not find chrome on the system")
			}
		} else if strings.ToLower(*rt.Browser.BrowserBinary) == "chromium" ||
			strings.ToLower(*rt.Browser.BrowserBinary) == "chromium-browser" {

			if _, err := os.Stat(DefaultOSXChromiumPath); err == nil && runtime.GOOS == "darwin" {
				return DefaultOSXChromiumPath, nil
			} else if _, err := os.Stat(DefaultLinuxChromiumPath); err == nil && runtime.GOOS == "linux" {
				return DefaultLinuxChromiumPath, nil
			} else if _, err := os.Stat(DefaultWindowsChromiumPath); err == nil && runtime.GOOS == "windows" {
				return DefaultWindowsChromiumPath, nil
			} else {
				return "", errors.New("could not find chrome on the system")
			}
		} else {
			return "", errors.New("could not find browser: " + *rt.Browser.BrowserBinary)
		}
	}
}

// getBrowserFlags uses the flag and extension settings passed in in the RawTask to create a single string
// slice with the flags we will use for our browser. Note that this slice will not include the specific
// flag which allows remote control of the browser. This flag will be added in Stage 3.
func getBrowserFlags(rt t.RawTask) ([]string, error) {
	result := make([]string, 0)

	if rt.Browser == nil {
		return DefaultChromiumBrowserFlags, nil
	}

	// We make copies of these two so we can manipulate them without altering the raw task
	abf := make([]string, 0)
	if rt.Browser.AddBrowserFlags != nil {
		abf = append(abf, *rt.Browser.AddBrowserFlags...)
	}
	rbf := make([]string, 0)
	if rt.Browser.RemoveBrowserFlags != nil {
		rbf = append(rbf, *rt.Browser.RemoveBrowserFlags...)
	}

	if rt.Browser.Extensions != nil && len(*rt.Browser.Extensions) != 0 {
		// Check that each extension exists
		for _, e := range *rt.Browser.Extensions {
			x, err := os.Stat(e)
			if err != nil {
				return []string{}, err
			}
			if !x.IsDir() {
				return []string{}, errors.New("given extension [ " + e + " ] is not a directory")
			}
		}

		// Create the extensions flag
		extensionsFlag := "--disable-extensions-except="
		extensionsFlag += (*rt.Browser.Extensions)[0]
		if len(*rt.Browser.Extensions) > 1 {
			for _, e := range (*rt.Browser.Extensions)[1:] {
				extensionsFlag += ","
				extensionsFlag += e
			}
		}

		abf = append(abf, extensionsFlag)

		// Remove the --incognito and --disable-extensions (both prevent extensions)
		rbf = append(rbf, "--incognito")
		rbf = append(rbf, "--disable-extensions")
	}

	if rt.Browser.SetBrowserFlags != nil && len(*rt.Browser.SetBrowserFlags) != 0 {
		if len(*rt.Browser.AddBrowserFlags) != 0 {
			log.Warn("SetBrowserFlags option is overriding non-empty AddBrowserFlags option.")
			log.Warn("Is this really what you intend?")
		}
		if len(*rt.Browser.RemoveBrowserFlags) != 0 {
			log.Warn("SetBrowserFlags option is overriding non-empty RemoveBrowserFlags option")
			log.Warn("Is this really what you intend?")
		}

		for _, flag := range *rt.Browser.SetBrowserFlags {
			result = append(result, flag)
		}
	} else {
		// Add flags, checking to see that they have not been removed
		for _, flag := range append(DefaultChromiumBrowserFlags, abf...) {
			for _, excluded := range rbf {
				if flag != excluded {
					result = append(result, flag)
				}
			}
		}
	}

	return result, nil
}

// getUserDataDirectory reads a raw task. If the task specifies a valid user data directory, it is
// returned. Otherwise, getUserDataDirectory selects a default directory based on the task UUID
func getUserDataDirectory(rt t.RawTask, uuid uuid.UUID) (string, error) {
	if rt.Browser != nil && rt.Browser.UserDataDirectory != nil && *rt.Browser.UserDataDirectory != "" {
		return *rt.Browser.UserDataDirectory, nil
	} else {
		// Use the first 8 characters of the uuid for temporary directories by default
		return path.Join(TempDir, uuid.String()[0:8]), nil
	}
}

// sanitizeCompletionSettings takes a raw CompletionSettings struct and sanitizes it
func sanitizeCompletionSettings(cs *t.CompletionSettings) (t.CompletionSettings, error) {
	result := task.AllocateNewCompletionSettings()

	if cs == nil {
		*result.CompletionCondition = DefaultCompletionCondition
		*result.Timeout = DefaultTimeout
		*result.TimeAfterLoad = DefaultTimeAfterLoad
		return *result, nil
	}

	if cs.CompletionCondition == nil {
		*result.CompletionCondition = DefaultCompletionCondition
	} else {
		for _, cc := range t.CompletionConditions {
			if cc == *cs.CompletionCondition {
				*result.CompletionCondition = *cs.CompletionCondition
			}
		}

		if *result.CompletionCondition == "" {
			return t.CompletionSettings{}, errors.New("invalid completion condition")
		}
	}

	if cs.Timeout == nil {
		*result.Timeout = DefaultTimeout
	} else {
		if *cs.Timeout <= 0 {
			return t.CompletionSettings{}, errors.New("timeout value must be positive")
		} else {
			*result.Timeout = *cs.Timeout
		}
	}

	if cs.TimeAfterLoad == nil {
		*result.TimeAfterLoad = DefaultTimeAfterLoad
	} else {
		if *cs.TimeAfterLoad < 0 {
			return t.CompletionSettings{}, errors.New("time_after_load value must be non-negative")
		} else {
			*result.TimeAfterLoad = *cs.TimeAfterLoad
		}
	}

	return *result, nil
}

// sanitizeDataSettings uses reflection to enumerate all DataSettings flags
// and sanitize/transfer them
func sanitizeDataSettings(ds *t.DataSettings, dds map[string]bool) (t.DataSettings, error) {
	result := task.AllocateNewDataSettings()

	fields := reflect.TypeOf(*result)
	values := reflect.ValueOf(*result)

	for i := 0; i < fields.NumField(); i += 1 {
		fieldName := fields.Field(i).Name
		val := values.Field(i).Interface().(*bool)

		if _, ok := DefaultDataSettings[fieldName]; !ok {
			return t.DataSettings{}, errors.New("found a data settings field not in DefaultDataSettings map")
		}

		*val = dds[fieldName]

		if ds != nil {
			dsFields := reflect.TypeOf(*ds)
			dsValues := reflect.ValueOf(*ds)

			for i := 0; i < dsFields.NumField(); i += 1 {
				dsFieldName := dsFields.Field(i).Name
				dsVal := dsValues.Field(i).Interface().(*bool)

				if dsVal != nil && dsFieldName == fieldName {
					*val = *dsVal
				}

			}
		}
	}

	return *result, nil
}
