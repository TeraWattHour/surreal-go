package surreal

import "github.com/terawatthour/surreal-go/rpc"

type Connection interface {
	Run()
	Send(method string, params []any) ([]byte, error)
	RegisterLiveCallback(id string, callback func(notification rpc.LiveNotification))
	Close() error
}
