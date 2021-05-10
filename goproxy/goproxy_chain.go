package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/cloudveiltech/goproxy"
)

func runConfigurationServerListener() {
	go func() {

		cert, _ := signHost(goproxy.GoproxyCa, []string{"127.0.0.1"})
		config := defaultTLSConfig
		config.Certificates = append(config.Certificates, *cert)
		config.NextProtos = []string{"http/1.1"}

		srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", configuredConfigurationServerPort)}
		srv.Handler = serverHandler{}
		srv.TLSConfig = config

		srv.ListenAndServeTLS("", "")
	}()
}

type serverHandler struct {
}

func (sh serverHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	localPort := fmt.Sprintf("%d", configuredConfigurationServerPort)
	remotePort := fmt.Sprintf("%d", configuredConfigurationServerPort+1)
	req.RequestURI = "http://127.0.0.1:" + remotePort + strings.ReplaceAll(req.RequestURI, localPort, remotePort)
	req.URL, _ = url.ParseRequestURI(req.RequestURI)
	req.Host = req.URL.Host

	log.Printf("Config server URI %s", req.RequestURI)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("Config server err %v", err)
		return
	}

	for k, vv := range resp.Header {
		if k != "Content-Length" {
			for _, v := range vv {
				log.Printf("Config server Response header %s:%s", k, v)
				rw.Header().Add(k, v)
			}
		}
	}
	rw.WriteHeader(resp.StatusCode)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	rw.Write(buf.Bytes())
	log.Printf("Config server Response sent %s", req.RequestURI)
}

func hashSorted(lst []string) []byte {
	c := make([]string, len(lst))
	copy(c, lst)
	sort.Strings(c)
	h := sha1.New()
	for _, s := range c {
		h.Write([]byte(s + ","))
	}
	return h.Sum(nil)
}

func hashSortedBigInt(lst []string) *big.Int {
	rv := new(big.Int)
	rv.SetBytes(hashSorted(lst))
	return rv
}

var goproxySignerVersion = ":goroxy1"

func signHost(ca tls.Certificate, hosts []string) (cert *tls.Certificate, err error) {
	var x509ca *x509.Certificate

	// Use the provided ca and not the global GoproxyCa for certificate generation.
	if x509ca, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return
	}
	start := time.Unix(0, 0)
	end, err := time.Parse("2006-01-02", "2049-12-31")
	if err != nil {
		panic(err)
	}

	serial := big.NewInt(rand.Int63())
	template := x509.Certificate{
		// TODO(elazar): instead of this ugly hack, just encode the certificate and hash the binary form.
		SerialNumber: serial,
		Issuer:       x509ca.Subject,
		Subject: pkix.Name{
			Organization: []string{"GoProxy untrusted MITM proxy Inc"},
		},
		NotBefore: start,
		NotAfter:  end,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
			template.Subject.CommonName = h
		}
	}

	hash := hashSorted(append(hosts, goproxySignerVersion, ":"+runtime.Version()))
	var csprng goproxy.CounterEncryptorRand
	if csprng, err = goproxy.NewCounterEncryptorRandFromKey(ca.PrivateKey, hash); err != nil {
		return
	}

	var certpriv crypto.Signer
	switch ca.PrivateKey.(type) {
	case *rsa.PrivateKey:
		if certpriv, err = rsa.GenerateKey(&csprng, 2048); err != nil {
			return
		}
	case *ecdsa.PrivateKey:
		if certpriv, err = ecdsa.GenerateKey(elliptic.P256(), &csprng); err != nil {
			return
		}
	default:
		err = fmt.Errorf("unsupported key type %T", ca.PrivateKey)
	}

	var derBytes []byte
	if derBytes, err = x509.CreateCertificate(&csprng, &template, x509ca, certpriv.Public(), ca.PrivateKey); err != nil {
		return
	}
	return &tls.Certificate{
		Certificate: [][]byte{derBytes, ca.Certificate[0]},
		PrivateKey:  certpriv,
	}, nil
}
