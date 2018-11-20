package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	goproxy "gopkg.in/elazarl/goproxy.v1"
)

var (
	caCert []byte = goproxy.CA_CERT
	caKey  []byte = goproxy.CA_KEY
)

func loadAndSetCa() {
	cert, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		return
	}
	key, err := ioutil.ReadFile("key.pem")
	if err != nil {
		return
	}

	caCert = cert
	caKey = key
	setCA(caCert, caKey)
}

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
