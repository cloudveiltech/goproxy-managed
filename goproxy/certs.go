package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/cloudveiltech/goproxy"
)

var (
	caCert []byte = goproxy.CA_CERT
	caKey  []byte = goproxy.CA_KEY
)

func loadAndSetCa(certFile, keyFile string) {
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Fatalf("Can't read cert file")
		return
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("Can't read cert key file")
		return
	}

	caCert = cert
	caKey = key
	setCA(caCert, caKey)
}

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		log.Fatalf("Can't load cert key/file")
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		log.Fatalf("Can't parse cert key/file")
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
