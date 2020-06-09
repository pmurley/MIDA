// This package contains most MIDA-native types. Many of these have members that come from types which
// need to be imported from other packages, but this package does not depend on any other MIDA packages.
package types

import (
	"github.com/chromedp/cdproto/network"
	"time"
)

// Settings describing the way in which a browser will be opened
type BrowserSettings struct {
	BrowserBinary      *string   `json:"browser_binary"`       // The binary for the browser (e.g., "/path/to/chrome.exe")
	UserDataDirectory  *string   `json:"user_data_directory"`  // Path to user data directory to use
	AddBrowserFlags    *[]string `json:"add_browser_flags"`    // Flags to be added to default browser flags
	RemoveBrowserFlags *[]string `json:"remove_browser_flags"` // Flags to be removed from default browser flags
	SetBrowserFlags    *[]string `json:"set_browser_flags"`    // Flags to use to override default browser flags
	Extensions         *[]string `json:"extensions"`           // Paths to browser extensions to be used for the crawl
}

// Conditions under which a crawl will complete successfully
type CompletionCondition string

const (
	TimeoutOnly   CompletionCondition = "TimeoutOnly"   // Complete only when the timeout is reached
	TimeAfterLoad CompletionCondition = "TimeAfterLoad" // Wait a given number of seconds after the load event
	LoadEvent     CompletionCondition = "LoadEvent"     // Terminate crawl immediately when load event fires
)

// Settings describing how a particular crawl will terminate
type CompletionSettings struct {
	CompletionCondition *CompletionCondition `json:"completion_condition"`      // Condition under which crawl will complete
	Timeout             *int                 `json:"timeout,omitempty"`         // Maximum amount of time the browser will remain open
	TimeAfterLoad       *int                 `json:"time_after_load,omitempty"` // Maximum amount of time the browser will remain open after page load
}

// Settings describing which data MIDA will capture from the crawl
type DataSettings struct {
	AllResources     *bool `json:"all_files"`         // Save all resource files
	ResourceMetadata *bool `json:"resource_metadata"` // Save extensive metadata about each resource
}

// Settings describing output of results to the local filesystem
type LocalOutputSettings struct {
	Path *string       `json:"path,omitempty"`          // Path over the overarching results directory to be written
	DS   *DataSettings `json:"data_settings,omitempty"` // Data settings for output to local filesystem
}

// Settings describing results output via SSH/SFTP
type SftpOutputSettings struct {
	Host           *string       `json:"host,omitempty"`          // IP address or domain name of host to store to
	Port           *int          `json:"port,omitempty"`          // Port to initiate SSH/SFTP connection
	Path           *string       `json:"path,omitempty"`          // Path of the overarching results directory to be written
	UserName       *string       `json:"user_name"`               // User name we should use for accessing the host
	PrivateKeyFile *string       `json:"private_key_file"`        // Path to the private key file we should use for accessing the host
	DS             *DataSettings `json:"data_settings,omitempty"` // Data settings for output via SSH/SFTP
}

// An aggregation of the output settings for a task or task-set
type OutputSettings struct {
	LocalOut *LocalOutputSettings `json:"local_output_settings"` // Output settings for the local filesystem
	SftpOut  *SftpOutputSettings  `json:"sftp_output_settings"`  // Output settings for the remote filesystem
}

// A raw MIDA task. This is the struct that is read from/written to file when tasks are stored as JSON.
type Task struct {
	URL *string `json:"url"` // The URL to be visited

	Browser    *BrowserSettings    `json:"browser_settings"`    // Settings for launching the browser
	Completion *CompletionSettings `json:"completion_settings"` // Settings for when the site visit will complete
	Data       *DataSettings       `json:"data_settings"`       // Settings for what data will be collected from the site
	Output     *OutputSettings     `json:"output_settings"`     // Settings for what/how results will be saved

	MaxAttempts *int `json:"max_attempts"` // Maximum number of failures before MIDA gives up on the task
	Repeat      *int `json:"repeat"`       // Number of times to repeat the crawl after it finishes successfully
}

// A slice of MIDA tasks, ready to be enqueued
type TaskSet []Task

// A grouping of tasks for multiple URLs with otherwise identical settings
type CompressedTaskSet struct {
	URL *[]string `json:"url"` // List of URLs to be visited

	Browser    *BrowserSettings    `json:"browser_settings"`    // Settings for launching the browser
	Completion *CompletionSettings `json:"completion_settings"` // Settings for when the site visit will complete
	Data       *DataSettings       `json:"data_settings"`       // Settings for what data will be collected from the site
	Output     *OutputSettings     `json:"output_settings"`     // Settings for what/how results will be saved

	MaxAttempts *int `json:"max_attempts"` // Maximum number of failures before MIDA gives up on the task
	Repeat      *int `json:"repeat"`       // Number of times to repeat the crawl after it finishes successfully
}

// Wrapper struct which contains a task, along with some dynamic metadata
type TaskWrapper struct {
	Task *Task // A pointer to a MIDA task

	CurrentAttempt   int      // The number of times we have tried this task so far
	FailureCode      string   // Holds the current failure code for the task, or "" if the task has not failed
	PastFailureCodes []string // Holds previous failures codes for the task
}

// Timing data for the processing of a particular task
type TaskTiming struct {
	BeginCrawl            time.Time `json:"begin_crawl"`
	BrowserOpen           time.Time `json:"browser_open"`
	ConnectionEstablished time.Time `json:"connection_established"`
	LoadEvent             time.Time `json:"load_event"`
	DOMContentEvent       time.Time `json:"dom_content_event"`
	BrowserClose          time.Time `json:"browser_close"`
	BeginPostprocess      time.Time `json:"begin_postprocess"`
	EndPostprocess        time.Time `json:"end_postprocess"`
	BeginStorage          time.Time `json:"begin_storage"`
	EndStorage            time.Time `json:"end_storage"`
}

// Statistics gathered about a specific task
type TaskSummary struct {
	Success     bool        `json:"success"`      // True if the task did not fail
	TaskWrapper TaskWrapper `json:"task_wrapper"` // Wrapper containing the full task
	TaskTiming  TaskTiming  `json:"task_timing"`  // Timing data for the task

	NumResources int `json:"num_resources,omitempty"` // Number of resources the browser loaded
}

// Information about the infrastructure used to perform the crawl
type CrawlerInfo struct {
	HostName    string `json:"host_name"`    // Host name of the machine used to crawl
	MidaVersion string `json:"mida_version"` // Version of MIDA used for this crawl

	Browser        string `json:"browser"`         // Name of the browser itself
	BrowserVersion string `json:"browser_version"` // Version of the browser we are using
	UserAgent      string `json:"user_agent"`      // User agent we are using
}

// The results MIDA gathers before they are post-processed
type RawResult struct {
	CrawlerInfo CrawlerInfo `json:"crawler_info"` // Information about the infrastructure used to crawl
	TaskWrapper TaskWrapper `json:"task_wrapper"` // Contains the full task that MIDA executed
}

type Resource struct {
	Requests  []network.EventRequestWillBeSent `json:"requests"`  // All requests sent for this particular request
	Responses []network.EventResponseReceived  `json:"responses"` // All responses received for this particular request
}

type FinalResult struct {
	Summary          TaskSummary         `json:"stats"`             // Statistics on timing and resource usage for the crawl
	ResourceMetadata map[string]Resource `json:"resource_metadata"` // Metadata on each resource loaded
}