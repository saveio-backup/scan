package dns

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (ds *Server) StartServer() {

}

// requestDns request dns from a full node
func (ds *Server) requestDns(paymentId, txId string, sig []byte) []byte {
	if !ds.checkPaymentValid(paymentId) {
		return nil
	}

	// request a full node
	return []byte{}
}

// checkPaymentValid check payment valid from nimbus
func (ds *Server) checkPaymentValid(paymentId string) bool {
	return true
}
