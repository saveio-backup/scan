package rest

const TLS_PORT int = 443

type ApiServer interface {
	Start() error
	Stop()
}
