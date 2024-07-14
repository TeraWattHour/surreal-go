package surreal

import (
	"fmt"
	"net/url"
)

type Options struct {
	Verbose          bool
	WebSocketOptions WebSocketOptions
}

func Connect(connectionUrl string, options *Options) (*DB, error) {
	parsedUrl, err := url.Parse(connectionUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid connection url: %s", err)
	}

	var conn Connection

	switch parsedUrl.Scheme {
	case "ws", "wss":
		conn, err = establishWebsocketConnection(connectionUrl, options)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported connection url scheme: %s", parsedUrl.Scheme)
	}

	go conn.Run()

	return &DB{
		conn:    conn,
		options: options,
	}, nil
}
