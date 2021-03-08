package src

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/Felix1Green/HttpProxyServer/internal/src/utils"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
)

type ProxyHandler struct {
	Cert *tls.Certificate
	Mu   *sync.RWMutex
}

func (t *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		t.serveSecureConnect(w, r)
	}

	proxy := httputil.ReverseProxy{}

	proxy.ServeHTTP(w, r)
}

func (t *ProxyHandler) serveSecureConnect(w http.ResponseWriter, r *http.Request) {
	dnsName, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		log.Println("incorrect request host:", r.Host)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	nameCert, err := utils.GenerateCert(t.Cert, []string{dnsName})
	if err != nil {
		log.Println("cannot create cert for name:", nameCert)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	serviceConfig := new(tls.Config)
	serviceConnection := new(tls.Conn)
	serviceConfig.Certificates = []tls.Certificate{*nameCert}
	serviceConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		clientConfig := new(tls.Config)
		clientConfig.ServerName = clientHello.ServerName
		serviceConnection, err = tls.Dial("tcp", clientConfig.ServerName, clientConfig)
		if err != nil {
			log.Println("cannot handle tls handshake with server:", clientHello.ServerName)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return nil, err
		}
		return utils.GenerateCert(t.Cert, []string{clientHello.ServerName})
	}

	clientConnection, err := utils.HandleHandshake(w, serviceConfig)
	if err != nil {
		log.Println("cannot handle tls handshake with client")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer func() {
		_ = clientConnection.Close()
		if serviceConnection != nil {
			_ = serviceConnection.Close()
		}
	}()

	if serviceConnection == nil {
		log.Println("service connection is nil")
		http.Error(w, "service connection cannot be established", http.StatusServiceUnavailable)
		return
	}

	proxy := httputil.ReverseProxy{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				t.Mu.Lock()
				defer t.Mu.Unlock()
				var c *tls.Conn = nil
				if serviceConnection == nil {
					return nil, errors.New("connection already used or even closed")
				}
				c = serviceConnection
				t.Mu.Unlock()
				return c, nil
			}},
	}

	proxy.ServeHTTP(w, r)
}
