package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/Felix1Green/HttpProxyServer/internal/src"
	"github.com/Felix1Green/HttpProxyServer/internal/src/utils"
	"log"
	"net/http"
	"sync"
)

func main() {
	config, _ := utils.GetConfig("")
	mutex := &sync.RWMutex{}
	cert, certErr := loadCert(config.CertFilePath, config.KeyFilePath)
	if certErr != nil {
		log.Fatalln(certErr)
	}
	pHandler := &src.ProxyHandler{
		Cert: &cert,
		Mu:   mutex,
	}

	log.Fatalln(http.ListenAndServe(config.ProxyPort, pHandler))
}

func loadCert(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil{
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	}
	return cert, err
}
