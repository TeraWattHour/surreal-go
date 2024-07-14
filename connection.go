package surreal

type Connection interface {
	Run()
	Send(method string, params []any) ([]byte, error)
	Close() error
}
