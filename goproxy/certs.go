package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/cloudveiltech/goproxy"
)

var defaultTLSConfig = &tls.Config{
	Renegotiation:      tls.RenegotiateFreelyAsClient,
	InsecureSkipVerify: true, // We should be able to set this to false, and then check verified chains against peer certificates to see if we have a trusted chain or not.
	VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// for i := 0; i < len(rawCerts); i++ {
		// 	cert, err := x509.ParseCertificate(rawCerts[i])

		// 	if err != nil {
		// 		fmt.Println("Error: ", err)
		// 		continue
		// 	}

		// 	hash := sha1.Sum(rawCerts[i])
		// 	fmt.Println("Cert data: ")
		// 	fmt.Println(hash, cert.DNSNames, cert.Subject, cert.Issuer)
		// }

		return nil
	},
}

var (
	caCert []byte = goproxy.CA_CERT
	caKey  []byte = goproxy.CA_KEY
)

func loadAndSetCa(certFile, keyFile string) {
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Printf("Can't read cert file")
		log.Fatal(err)
		return
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Printf("Can't read cert key file")
		log.Fatal(err)
		return
	}

	caCert = cert
	caKey = key
	setCA(caCert, caKey)
}

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		log.Printf("Can't load cert/key file")
		log.Fatal(err)
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		log.Printf("Can't parse cert key/file")
		log.Fatal(err)
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}

func verifyCerts(dnsName string, peerCerts []*x509.Certificate) (bool, error) {
	dnsNamePatched := strings.Split(dnsName, ":")[0]

	opts := x509.VerifyOptions{
		Roots:         nil,
		DNSName:       dnsNamePatched,
		Intermediates: x509.NewCertPool(),
		CurrentTime:   time.Now(),
	}

	for i, cert := range peerCerts {
		if i == 0 {
			continue
		}

		opts.Intermediates.AddCert(cert)
	}

	var err error
	_, err = peerCerts[0].Verify(opts)
	if err != nil {
		log.Printf("Verify certs error %v", err)
		return false, err
	}

	return true, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func GenerateCerts(caCertPath, caKeyPath string) bool {

	// priv, err := rsa.GenerateKey(rand.Reader, *rsaBits)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("Error generating cert %s", err)
		return false
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Cloudveil Filtering Certificate"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 3650),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Printf("Failed to create certificate: %s", err)
		return false
	}

	certFile, err := os.Create(caCertPath)
	if err != nil {
		log.Printf("Error generating cert %s", err)
		return false
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	certKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		log.Printf("Unable to marshal ECDSA private key: %v", err)
		return false
	}
	defer certKeyFile.Close()
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Printf("Unable to marshal ECDSA private key: %v", err)
		return false
	}
	if err := pem.Encode(certKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}); err != nil {
		log.Printf("Unable to marshal ECDSA private key: %v", err)
		return false
	}

	return true
}
