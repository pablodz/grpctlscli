package grpctls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

func fetchCertificate(host string) (*x509.Certificate, error) {
	conn, err := tls.Dial("tcp", host, &tls.Config{})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	return certs[0], nil
}
