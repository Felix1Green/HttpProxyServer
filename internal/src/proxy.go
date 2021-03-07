package src

import (
	"crypto/tls"
	"net/http"
)

type ProxyHandler struct{
	Cert *tls.Certificate
}


func (t *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){

}