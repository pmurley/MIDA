package main

import (
	t "github.com/pmurley/mida/types"
	"github.com/spf13/viper"
	"os"
	"path"
)

// initViperConfig
func initViperConfig() {
	// Initialize the hardcoded defaults
	setDefaults()

	// We will read environment variables with the "MIDA" prefix
	viper.SetEnvPrefix("MIDA")
	viper.AutomaticEnv()
}

// Hardcoded default configuration values
func setDefaults() {
	// MIDA-Wide Configuration Defaults
	viper.SetDefault("crawlers", 1)
	viper.SetDefault("storers", 1)
	viper.SetDefault("prom-port", 8001)
	viper.SetDefault("monitor", false)
	viper.SetDefault("log-level", 2)
	viper.SetDefault("task-file", "examples/example_task.json")
	viper.SetDefault("rabbitmqurl", "localhost:5672")
	viper.SetDefault("rabbitmquser", "")
	viper.SetDefault("rabbitmqpass", "")
	viper.SetDefault("rabbitmqtaskqueue", "tasks")
	viper.SetDefault("rabbitmqbroadcastqueue", "broadcast")
}

const (
	// MIDA Configuration Defaults

	DefaultTaskAttempts         = 1  // How many times to try a task before we give up and fail
	DefaultNavTimeout           = 30 // How long to wait when connecting to a web server
	DefaultSSHBackoffMultiplier = 5  // Exponential increase in time between tries when connecting for SFTP storage
	DefaultTaskPriority         = 5  // Queue priority when creating new tasks -- Value should be 1-10

	// Browser-Related Parameters
	DefaultOSXChromePath       = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	DefaultOSXChromiumPath     = "/Applications/Chromium.app/Contents/MacOS/Chromium"
	DefaultLinuxChromePath     = "/usr/bin/google-chrome-stable"
	DefaultLinuxChromiumPath   = "/usr/bin/chromium-browser"
	DefaultWindowsChromePath   = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
	DefaultWindowsChromiumPath = "\\%LocalAppData%\\chromium\\Application\\chrome.exe"

	// Task completion
	DefaultTimeAfterLoad       = 0  // Default time to stay on a page after load event is fired (in TimeAfterLoad mode)
	DefaultTimeout             = 10 // Default time (in seconds) to remain on a page before exiting browser
	DefaultCompletionCondition = t.TimeoutOnly

	// Defaults for data gathering settings
	DefaultAllResources     = true
	DefaultResourceMetadata = true

	DefaultShuffle = true // Whether to shuffle order of task processing

	AlphaNumChars           = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890" // Set of characters for generating unique IDs
	DefaultIdentifierLength = 16                                     // Random identifier for each crawl
	DefaultProtocolPrefix   = "https://"                             // If no protocol is provided, we use https for the crawl
)

var (
	TempDir = path.Join(os.TempDir(), "MIDA") // Directory MIDA will use for temporary files
)

// Flags we apply by default to Chrome/Chromium-based browsers
var DefaultChromiumBrowserFlags = []string{
	"--enable-features=NetworkService",
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-client-side-phishing-detection",
	"--disable-extensions",
	"--disable-features=IsolateOrigins,site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-infobars",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disk-cache-size=0",
	"--incognito",
	"--new-window",
	"--no-default-browser-check",
	"--no-first-run",
	"--no-sandbox",
	"--safebrowsing-disable-auto-update",
}
