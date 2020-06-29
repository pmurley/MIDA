package browser

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	b "github.com/pmurley/mida/base"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

func VisitPageDevtoolsProtocol(tw b.TaskWrapper) (*b.RawResult, error) {

	// Fully allocate our raw result object -- should be locked whenever it is read or written
	rawResult := b.RawResult{
		CrawlerInfo: b.CrawlerInfo{},
		TaskSummary: b.TaskSummary{
			Success:      false,
			TaskWrapper:  tw,
			TaskTiming:   b.TaskTiming{},
			NumResources: 0,
		},
	}

	// Open all the event channels we will use to receive events from DevTools
	ec := openEventChannels()

	// Make sure user data directory exists already. If not, create it.
	// If we can't create it, we consider it a bad enough error that we
	// return an error and stop MIDA entirely -- likely a major misconfiguration
	_, err := os.Stat(tw.SanitizedTask.UserDataDirectory)
	if err != nil {
		err = os.MkdirAll(tw.SanitizedTask.UserDataDirectory, 0744)
		if err != nil {
			return nil, err
		}
	}

	// If we are gathering all the resources, we need to create the corresponding directory
	if *(tw.SanitizedTask.DS.AllResources) {
		// Create a subdirectory where we will store all the files
		_, err = os.Stat(path.Join(tw.SanitizedTask.UserDataDirectory, b.DefaultResourceSubdir))
		if err != nil {
			err = os.MkdirAll(path.Join(tw.SanitizedTask.UserDataDirectory, b.DefaultResourceSubdir), 0744)
			if err != nil {
				return nil, err
			}
		}
	}

	// Build our opts slice
	var opts []chromedp.ExecAllocatorOption
	for _, flagString := range tw.SanitizedTask.BrowserFlags {
		name, val, err := ChromeFormatFlag(flagString)
		if err != nil {
			// We got a bad flag
			continue
		}
		opts = append(opts, chromedp.Flag(name, val))
	}

	opts = append(opts, chromedp.Flag("user-data-dir", tw.SanitizedTask.UserDataDirectory))
	opts = append(opts, chromedp.ExecPath(tw.SanitizedTask.BrowserBinaryPath))

	// Build channels we need for coordinating the site visit across goroutines
	navChan := make(chan error)                                                          // A channel to signal the completion of navigation, successfully or not
	timeoutChan := time.After(time.Duration(*tw.SanitizedTask.CS.Timeout) * time.Second) // Absolute longest we can go
	loadEventChan := make(chan bool)                                                     // Used to signal the firing of load events
	var eventHandlerWG sync.WaitGroup                                                    // Used to make sure all the event handlers exit

	// Get our event listener goroutines up and running
	eventListenerContext, eventListenerCancel := context.WithCancel(context.Background())
	eventHandlerWG.Add(2) // UPDATE ME WHEN YOU ADD A NEW EVENT HANDLER
	go PageLoadEventFired(ec.loadEventFiredChan, loadEventChan, &rawResult, &eventHandlerWG, eventListenerContext)
	go NetworkRequestWillBeSent(ec.requestWillBeSentChan, &rawResult, &eventHandlerWG, eventListenerContext)

	// Spawn our browser
	allocContext, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	browserContext, _ := chromedp.NewContext(allocContext)

	// Ensure the correct domains are enabled/disabled
	err = chromedp.Run(browserContext, chromedp.ActionFunc(func(cxt context.Context) error {
		err = runtime.Disable().Do(cxt)
		if err != nil {
			return err
		}

		err = page.Enable().Do(cxt)
		if err != nil {
			return err
		}

		_, err = debugger.Enable().Do(cxt)
		if err != nil {
			return err
		}

		err = network.Enable().Do(cxt)
		if err != nil {
			return err
		}

		return nil
	}))

	// Event Demux - just receive the events and stick them in the applicable channels
	chromedp.ListenTarget(browserContext, func(ev interface{}) {
		switch ev.(type) {
		case *page.EventLoadEventFired:
			ec.loadEventFiredChan <- ev.(*page.EventLoadEventFired)
		case *network.EventRequestWillBeSent:
			ec.requestWillBeSentChan <- ev.(*network.EventRequestWillBeSent)
		}
	})

	// Initiate navigation to the applicable page
	go func() {
		err = chromedp.Run(browserContext, chromedp.ActionFunc(func(ctxt context.Context) error {
			_, _, text, err := page.Navigate(tw.SanitizedTask.URL).Do(ctxt)
			if err != nil {
				return err
			} else if text != "" {
				return errors.New(text)
			} else {
				return nil
			}
		}))
		navChan <- err
	}()

	select {
	case err = <-navChan:
		rawResult.Lock()
		rawResult.TaskSummary.TaskTiming.ConnectionEstablished = time.Now()
		rawResult.Unlock()
	case <-time.After(b.DefaultNavTimeout * time.Second):
		// Our connection to the web server took longer than out navigation timeout (currently 30 seconds)
		err = errors.New("timeout on connection to webserver")
	case <-timeoutChan:
		err = errors.New("total site visit time exceeded before we connected to webserver")
	case <-browserContext.Done():
		// The browser somehow closed before we finished navigation
		err = errors.New("browser closed during connection to site")
	}
	if err != nil {
		// Save our error message for storage
		errorCode := err.Error()

		// We have failed to navigate to the site. Shut things down.
		closeContext, _ := context.WithTimeout(browserContext, 5*time.Second)
		err = chromedp.Cancel(closeContext)
		if err != nil {
			// We failed to close chrome gracefully within the allotted timeout
		}

		// Close all of our event channels and then wait for any outstanding event handlers to finish
		closeEventChannels(ec)
		eventListenerCancel()

		eventHandlerWG.Wait()

		rawResult.Lock()
		rawResult.TaskSummary.TaskWrapper.FailureCode = errorCode
		rawResult.TaskSummary.Success = false
		rawResult.TaskSummary.TaskTiming.BrowserClose = time.Now()
		rawResult.Unlock()

		return &rawResult, nil
	}

	// We have now successfully connected and navigated to the site

	select {
	case <-browserContext.Done():
		// Browser crashed, closed manually, or we otherwise lost connection to it prematurely
	case <-loadEventChan:
		// The load event fired. What we do next depends on how the crawl completes
		switch *tw.SanitizedTask.CS.CompletionCondition {
		case b.TimeAfterLoad:
			select {
			case <-browserContext.Done():
				// Browser crashed, closed manually, or we otherwise lost connection to it prematurely
			case <-timeoutChan:
				// We hit our general timeout before we got to timeAfterLoad. Fall through to browser close and cleanup
			case <-time.After(time.Duration(*tw.SanitizedTask.CS.TimeAfterLoad) * time.Second):
				// We finished our timeAfterLoad period. Fall through to browser close and cleanup
			}
		case b.LoadEvent:
			// We got out load event so we are just done, fall through to browser close and cleanup
		case b.TimeoutOnly:
			// We need to just continue waiting for the timeout (or unexpected browser close)
			select {
			case <-browserContext.Done():
				// Browser crashed, closed manually, or we otherwise lost connection to it prematurely
			case <-timeoutChan:
				// We hit our general timeout, fall through to browser close and cleanup
			}
		default:
			// This state should be unreachable -- got an unknown termination condition
		}
	}

	closeContext, _ := context.WithTimeout(browserContext, 5*time.Second)
	err = chromedp.Cancel(closeContext)
	if err != nil {
		// This isn't an ideal solution, but if the graceful close fails, we have to just kill the browser to free resources
		allocCancel()
	}

	// Store time at which we closed the browser
	rawResult.Lock()
	rawResult.TaskSummary.TaskTiming.BrowserClose = time.Now()
	rawResult.Unlock()

	// Signal to shut down all event listeners
	eventListenerCancel()

	// Wait for all event handlers to finish
	fmt.Println("before")
	eventHandlerWG.Wait()
	fmt.Println("after")

	return &rawResult, nil
}

type EventChannels struct {
	loadEventFiredChan                     chan *page.EventLoadEventFired
	domContentEventFiredChan               chan *page.EventDomContentEventFired
	requestWillBeSentChan                  chan *network.EventRequestWillBeSent
	responseReceivedChan                   chan *network.EventResponseReceived
	loadingFinishedChan                    chan *network.EventLoadingFinished
	dataReceivedChan                       chan *network.EventDataReceived
	webSocketCreatedChan                   chan *network.EventWebSocketCreated
	webSocketFrameSentChan                 chan *network.EventWebSocketFrameSent
	webSocketFrameReceivedChan             chan *network.EventWebSocketFrameReceived
	webSocketFrameErrorChan                chan *network.EventWebSocketFrameError
	webSocketClosedChan                    chan *network.EventWebSocketClosed
	webSocketWillSendHandshakeRequestChan  chan *network.EventWebSocketWillSendHandshakeRequest
	webSocketHandshakeResponseReceivedChan chan *network.EventWebSocketHandshakeResponseReceived
	EventSourceMessageReceivedChan         chan *network.EventEventSourceMessageReceived
	requestPausedChan                      chan *fetch.EventRequestPaused
	scriptParsedChan                       chan *debugger.EventScriptParsed
}

func openEventChannels() EventChannels {
	ec := EventChannels{
		loadEventFiredChan:                     make(chan *page.EventLoadEventFired, b.DefaultEventChannelBufferSize),
		domContentEventFiredChan:               make(chan *page.EventDomContentEventFired, b.DefaultEventChannelBufferSize),
		requestWillBeSentChan:                  make(chan *network.EventRequestWillBeSent, b.DefaultEventChannelBufferSize),
		responseReceivedChan:                   make(chan *network.EventResponseReceived, b.DefaultEventChannelBufferSize),
		loadingFinishedChan:                    make(chan *network.EventLoadingFinished, b.DefaultEventChannelBufferSize),
		dataReceivedChan:                       make(chan *network.EventDataReceived, b.DefaultEventChannelBufferSize),
		webSocketCreatedChan:                   make(chan *network.EventWebSocketCreated, b.DefaultEventChannelBufferSize),
		webSocketFrameSentChan:                 make(chan *network.EventWebSocketFrameSent, b.DefaultEventChannelBufferSize),
		webSocketFrameReceivedChan:             make(chan *network.EventWebSocketFrameReceived, b.DefaultEventChannelBufferSize),
		webSocketFrameErrorChan:                make(chan *network.EventWebSocketFrameError, b.DefaultEventChannelBufferSize),
		webSocketClosedChan:                    make(chan *network.EventWebSocketClosed, b.DefaultEventChannelBufferSize),
		webSocketWillSendHandshakeRequestChan:  make(chan *network.EventWebSocketWillSendHandshakeRequest, b.DefaultEventChannelBufferSize),
		webSocketHandshakeResponseReceivedChan: make(chan *network.EventWebSocketHandshakeResponseReceived, b.DefaultEventChannelBufferSize),
		EventSourceMessageReceivedChan:         make(chan *network.EventEventSourceMessageReceived, b.DefaultEventChannelBufferSize),
		requestPausedChan:                      make(chan *fetch.EventRequestPaused, b.DefaultEventChannelBufferSize),
		scriptParsedChan:                       make(chan *debugger.EventScriptParsed, b.DefaultEventChannelBufferSize),
	}

	return ec
}

func closeEventChannels(ec EventChannels) {
	close(ec.loadEventFiredChan)
	close(ec.domContentEventFiredChan)
	close(ec.requestWillBeSentChan)
	close(ec.responseReceivedChan)
	close(ec.loadingFinishedChan)
	close(ec.dataReceivedChan)
	close(ec.webSocketCreatedChan)
	close(ec.webSocketFrameSentChan)
	close(ec.webSocketFrameReceivedChan)
	close(ec.webSocketFrameErrorChan)
	close(ec.webSocketClosedChan)
	close(ec.webSocketWillSendHandshakeRequestChan)
	close(ec.webSocketHandshakeResponseReceivedChan)
	close(ec.EventSourceMessageReceivedChan)
	close(ec.requestPausedChan)
	close(ec.scriptParsedChan)
}

// ChromeFormatFlag takes a variety of possible flag formats and puts them in a format that chromedp understands (key/value)
func ChromeFormatFlag(f string) (string, interface{}, error) {
	if strings.HasPrefix(f, "--") {
		f = f[2:]
	}

	parts := strings.Split(f, "=")
	if len(parts) == 1 {
		return parts[0], true, nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], nil
	}

	return "", "", errors.New("invalid flag: " + f)
}

func PageLoadEventFired(eventChan chan *page.EventLoadEventFired, loadEventChan chan<- bool, rawResult *b.RawResult, wg *sync.WaitGroup, ctxt context.Context) {
	done := false
	for {
		select {
		case _, ok := <-eventChan:
			if !ok {
				// Channel closed
				done = true
			}

			rawResult.Lock()
			rawResult.TaskSummary.TaskTiming.LoadEvent = time.Now()
			rawResult.Unlock()

			// Signal that a load event has fired
			loadEventChan <- true

		case <-ctxt.Done():
			// Context canceled
			done = true
		}

		if done {
			break
		}
	}

	fmt.Println("page")
	wg.Done()
}

func NetworkRequestWillBeSent(eventChan chan *network.EventRequestWillBeSent, rawResult *b.RawResult, wg *sync.WaitGroup, ctxt context.Context) {

	done := false
	for {
		select {
		case ev, ok := <-eventChan:
			if !ok {
				// Channel closed
				done = true
			}

			rawResult.Lock()
			//spew.Dump(ev)
			_ = ev.Request.URL
			rawResult.Unlock()
		case <-ctxt.Done():
			// Context canceled
			done = true
		}

		if done {
			break
		}
	}

	fmt.Println("network")
	wg.Done()
}
