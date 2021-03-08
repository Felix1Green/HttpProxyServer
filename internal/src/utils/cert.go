package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"log"
	"math/big"
	"net/http"
	"time"
)

const (
	caMaxAge   = 5 * 365 * 24 * time.Hour
	leafMaxAge = 24 * time.Hour
	leafUsage  = x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
)

var (
	connectionEstablishedResponse = []byte("HTTP/1.1 200 OK\r\n\r\n")
)

func GenerateCert(c *tls.Certificate, names []string) (*tls.Certificate, error) {
	if !c.Leaf.IsCA {
		return nil, errors.New("certificate is not CA")
	}

	currentTime := time.Now().UTC()
	serialNumber, randErr := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if randErr != nil {
		return nil, errors.New("cant create random serial number for certificate")
	}
	tmpCert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: names[0]},
		NotBefore:             currentTime,
		NotAfter:              currentTime.Add(leafMaxAge),
		KeyUsage:              leafUsage,
		BasicConstraintsValid: true,
		DNSNames:              names,
		SignatureAlgorithm:    x509.SHA256WithRSA,
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rawCert, err := x509.CreateCertificate(rand.Reader, tmpCert, c.Leaf, key.Public(), c.PrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	cert := new(tls.Certificate)
	cert.PrivateKey = c.PrivateKey
	cert.Certificate = append(cert.Certificate, rawCert)
	cert.Leaf, _ = x509.ParseCertificate(rawCert)
	return cert, nil
}

func HandleHandshake(w http.ResponseWriter, config *tls.Config) (*tls.Conn, error) {
	rawConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, err
	}

	if _, err := rawConn.Write(connectionEstablishedResponse); err != nil {
		_ = rawConn.Close()
		log.Println("cannot write connection established response to client")
		return nil, err
	}
	clientConn := tls.Server(rawConn, config)
	log.Println("before hand")
	err = clientConn.Handshake()
	if err != nil {
		_ = clientConn.Close()
		_ = rawConn.Close()
		return nil, err
	}
	log.Println("after hand")
	return clientConn, nil
}
