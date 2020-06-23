package browser

import (
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/davecgh/go-spew/spew"
	b "github.com/pmurley/mida/base"
	"os"
	"path"
)

func VisitPageDevtoolsProtocol(tw b.TaskWrapper) (b.RawResult, error) {

	// Fully allocate our raw result object -- should be locked whenever it is read or written
	/*
		rawResult := b.RawResult{
			CrawlerInfo: b.CrawlerInfo{},
			TaskWrapper: tw,
		}
	*/
	// Open all the event channels we will use to receive events from DevTools
	/*
		ec := openEventChannels()
	*/
	// Make sure user data directory exists already. If not, create it.
	// If we can't create it, we consider it a bad enough error that we
	// return an error and stop MIDA entirely -- likely a major misconfiguration
	_, err := os.Stat(tw.SanitizedTask.UserDataDirectory)
	if err != nil {
		err = os.MkdirAll(tw.SanitizedTask.UserDataDirectory, 0744)
		if err != nil {
			return b.RawResult{}, err
		}
	}

	// If we are gathering all the resources, we need to create the corresponding directory
	if *(tw.SanitizedTask.DS.AllResources) {
		// Create a subdirectory where we will store all the files
		_, err = os.Stat(path.Join(tw.SanitizedTask.UserDataDirectory, b.DefaultResourceSubdir))
		if err != nil {
			err = os.MkdirAll(path.Join(tw.SanitizedTask.UserDataDirectory, b.DefaultResourceSubdir), 0744)
			if err != nil {
				return b.RawResult{}, err
			}
		}
	}

	spew.Dump(tw)
	return b.RawResult{}, nil
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
