package storage

import (
	"encoding/json"
	"errors"
	b "github.com/pmurley/mida/base"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func Store(finalResult b.FinalResult) error {
	var err error

	// For brevity
	st := finalResult.Summary.TaskWrapper.SanitizedTask

	if *st.OPS.LocalOut.Enable {
		err = Local(finalResult)
		if err != nil {
			return err
		}
	}

	return nil
}

func Local(finalResult b.FinalResult) error {

	// For brevity
	st := &finalResult.Summary.TaskWrapper.SanitizedTask
	tw := &finalResult.Summary.TaskWrapper

	// Build our output path
	dirName, err := DirNameFromURL(st.URL)
	if err != nil {
		return errors.New("failed to extract directory name from URL: " + err.Error())
	}
	outPath := path.Join(*st.OPS.LocalOut.Path, dirName, finalResult.Summary.TaskWrapper.UUID.String())

	_, err = os.Stat(outPath)
	if err != nil {
		err = os.MkdirAll(outPath, 0755)
		if err != nil {
			return errors.New("failed to create local output directory: " + err.Error())
		}
	} else {
		return errors.New("task local output directory exists")
	}

	if *st.OPS.LocalOut.DS.ResourceMetadata {
		data, err := json.Marshal(finalResult.DTResourceMetadata)
		if err != nil {
			return errors.New("failed to marshal resource data for local storage: " + err.Error())
		}

		err = ioutil.WriteFile(path.Join(outPath, b.DefaultResourceMetadataFile), data, 0644)
		if err != nil {
			return errors.New("failed to write resource metadata file: " + err.Error())
		}
	}

	if *st.OPS.LocalOut.DS.AllResources {
		err = os.Rename(path.Join(tw.TempDir, b.DefaultResourceSubdir), path.Join(outPath, b.DefaultResourceSubdir))
		if err != nil {
			return errors.New("failed to copy resources directory into results directory")
		}
	}

	// Chrome sometimes won't allow the user data directory to be deleted on the first try,
	// so we loop until we can successfully remove it
	err = os.RemoveAll(tw.SanitizedTask.UserDataDirectory)
	for {
		if _, err := os.Stat(tw.SanitizedTask.UserDataDirectory); err == nil {
			time.Sleep(1 * time.Second)
			err = os.RemoveAll(tw.SanitizedTask.UserDataDirectory)
		} else {
			break
		}
	}

	// Store our log
	tw.LogFile.Close()
	err = os.Rename(tw.LogFile.Name(), path.Join(outPath, b.DefaultTaskLogFile))
	if err != nil {
		// We failed to write the log file -- WHERE DO WE LOG THIS   :/
	}

	// Remove our temporary results directory
	err = os.RemoveAll(tw.TempDir)
	if err != nil {
		// We failed to remove our temporary results directory
	}

	return nil
}

// DirNameFromURL takes a URL and sanitizes/escapes it so it can safely be used as a filename
func DirNameFromURL(s string) (string, error) {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return "", err
	}

	// Replace all disallowed file path characters (both Windows and Unix) so we can safely use URL as directory name
	disallowedChars := []string{"/", "\\", ">", "<", ":", "|", "?", "*"}
	result := u.Host + u.EscapedPath()
	for _, c := range disallowedChars {
		result = strings.Replace(result, c, "-", -1)
	}
	return result, nil
}
