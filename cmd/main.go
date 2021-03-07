package main

import(
	"crypto/tls"
	"net/http"
	"github.com/Felix1Green/HttpProxyServer/internal/src"
	"github.com/Felix1Green/HttpProxyServer/internal/src/utils"
	"log"
)


func main(){
	config, _ := utils.GetConfig("")
	cert, certErr := loadCert(config.CertFilePath, config.KeyFilePath)
	if certErr != nil{
		log.Fatalln(certErr)
	}
	pHandler := &src.ProxyHandler{
		Cert: &cert,
	}

	log.Fatalln(http.ListenAndServe(config.Port, pHandler))
}


func loadCert(certFile, keyFile string) (tls.Certificate, error){
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	return cert, err
}