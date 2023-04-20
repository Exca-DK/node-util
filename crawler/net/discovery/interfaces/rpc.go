package interfaces

type RpcInfo interface {
	Error() error
	Packet() string
}

func WrapIntoRpcInfo(err error, _type string) RpcInfo {
	return rpcInfo{err: err, _type: _type}
}

type rpcInfo struct {
	err   error
	_type string
}

func (err rpcInfo) Error() error   { return err.err }
func (err rpcInfo) Packet() string { return err._type }
