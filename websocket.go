package surreal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/terawatthour/surreal-go/rpc"
	"log"
	"sync"
	"time"
)

const (
	Alphanumeric   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	DefaultTimeout = 10 * time.Second
)

type WebSocketOptions struct {
	// DisableCompression enables compression for the websocket connection.
	DisableCompression bool

	// OnDropCallback is called when the websocket connection is dropped.
	OnDropCallback func(reason error)

	// ResponseTimeout is the duration to wait for a response before timing out. Defaults to 10 seconds.
	ResponseTimeout time.Duration
}

func (o *WebSocketOptions) responseTimeout() time.Duration {
	if o.ResponseTimeout == 0 {
		return DefaultTimeout
	}
	return o.ResponseTimeout
}

type WebSocketConnection struct {
	options *Options

	conn     *websocket.Conn
	connLock sync.Mutex

	responseChannels     map[string]chan rpc.Incoming
	responseChannelsLock sync.RWMutex

	done     chan struct{}
	doneOnce sync.Once
}

func establishWebsocketConnection(url string, options *Options) (Connection, error) {
	if options == nil {
		options = &Options{}
	}

	dialer := websocket.DefaultDialer
	dialer.EnableCompression = !options.WebSocketOptions.DisableCompression
	dialer.HandshakeTimeout = 10 * time.Second

	c, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %s", err)
	}

	conn := &WebSocketConnection{
		conn:             c,
		options:          options,
		done:             make(chan struct{}),
		responseChannels: make(map[string]chan rpc.Incoming),
	}

	return conn, nil
}

// Send writes a message to the websocket connection and waits for a response.
// Expects a JSON serializable object.
func (ws *WebSocketConnection) Send(method string, params []any) ([]byte, error) {
	select {
	case <-ws.done:
		return nil, fmt.Errorf("connection is closed")
	default:
	}

	eventId, _ := gonanoid.Generate(Alphanumeric, 16)
	outgoing := &rpc.Outgoing{
		ID:     eventId,
		Method: method,
		Params: params,
	}

	ch := ws.openResponseChannel(eventId)
	defer ws.removeResponseChannel(eventId)

	if err := ws.write(outgoing); err != nil {
		return nil, fmt.Errorf("failed to write message to websocket: %s", err)
	}

	timeout := time.After(ws.options.WebSocketOptions.responseTimeout())
	select {
	case <-timeout:
		return nil, fmt.Errorf("request timed out")
	case <-ws.done:
		return nil, fmt.Errorf("connection dropped before response was received")
	case res, open := <-ch:
		if !open {
			return nil, fmt.Errorf("response channel closed before response was received")
		}
		if res.Error != nil {
			return nil, res.Error
		}
		return res.Result, nil
	}
}

func (ws *WebSocketConnection) Run() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ws.done:
			return
		case <-ticker.C:
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				_ = ws.close(err)

				if ws.options.Verbose {
					log.Println("failed to send ping to websocket: ", err)
				}
				return
			}
		default:
			var incoming rpc.Incoming
			_, msg, err := ws.conn.ReadMessage()
			if err != nil {
				_ = ws.close(err)

				if ws.options.Verbose {
					log.Println("failed to read message from websocket: ", err)
				}
				return
			}

			if err := json.Unmarshal(msg, &incoming); err != nil {
				if ws.options.Verbose {
					log.Println("failed to unmarshal message from surreal: ", err)
				}
				continue
			}

			go ws.handleResponse(incoming)
		}
	}
}

func (ws *WebSocketConnection) Close() error {
	return ws.close(nil)
}

func (ws *WebSocketConnection) write(msg any) error {
	marshalled, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ws.connLock.Lock()
	defer ws.connLock.Unlock()

	return ws.conn.WriteMessage(websocket.TextMessage, marshalled)
}

func (ws *WebSocketConnection) close(reason error) error {
	defer func() {
		ws.doneOnce.Do(func() {
			close(ws.done)
		})

		ws.connLock.Unlock()
		if reason != nil && ws.options != nil && ws.options.WebSocketOptions.OnDropCallback != nil {
			ws.options.WebSocketOptions.OnDropCallback(reason)
		}
	}()

	ws.connLock.Lock()
	_ = ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, ""))

	return ws.conn.Close()
}

func (ws *WebSocketConnection) handleResponse(incoming rpc.Incoming) {
	switch incoming.ID {
	case "", nil:
		panic("live notifications and events are not supported")
	default:
		ch, ok := ws.acquireResponseChannel(fmt.Sprintf("%v", incoming.ID))
		if !ok {
			return
		}
		ch <- incoming
		close(ch)
	}
}

func (ws *WebSocketConnection) openResponseChannel(eventId string) chan rpc.Incoming {
	ws.responseChannelsLock.Lock()
	defer ws.responseChannelsLock.Unlock()

	ch := make(chan rpc.Incoming)
	ws.responseChannels[eventId] = ch

	return ch
}

func (ws *WebSocketConnection) acquireResponseChannel(eventId string) (chan rpc.Incoming, bool) {
	ws.responseChannelsLock.RLock()
	defer ws.responseChannelsLock.RUnlock()

	ch, ok := ws.responseChannels[eventId]
	return ch, ok
}

func (ws *WebSocketConnection) removeResponseChannel(eventId string) {
	ws.responseChannelsLock.Lock()
	defer ws.responseChannelsLock.Unlock()

	delete(ws.responseChannels, eventId)
}
