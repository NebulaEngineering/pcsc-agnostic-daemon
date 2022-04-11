package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func verifyAndCreateFiles(certName, keyName string, create bool) (string, string, error) {
	cert := certName
	key := keyName

	if isDir(certName) {
		cert = cert + filenameCert
	}
	if len(keyName) <= 0 {
		key = filepath.Dir(cert) + filenameKey
	} else if isDir(keyName) {
		key = key + filenameKey

	}
	switch {
	case create && len(certName) <= 0:
		return "", "", fmt.Errorf("option \"create\" is \"true\" but \"certpath\" is not defined")
	case len(cert) > 0 && len(key) > 0 && !create:
		return cert, key, nil
	case len(cert) > 0 && len(key) > 0 && fileExists(cert):
		return cert, key, nil
	case len(cert) > 0:
		if len(key) > 0 && fileExists(key) {
			return "", "", fmt.Errorf("if \"keypath\" (%s) is defined, \"certpath\" (%s) could be a file in filesystem", key, cert)
		}
		if !pathExists(cert) {
			return "", "", fmt.Errorf("\"%s\" isn't a valid file in filesystem", cert)
		}

		fcert, err := os.OpenFile(cert, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return "", "", err
		}
		fkey, err := os.OpenFile(key, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return "", "", err
		}
		certData, keyData, err := newCert()
		if err != nil {
			return "", "", err
		}
		if err := writeFile(fcert, certData); err != nil {
			return "", "", err
		}
		if err := writeFile(fkey, keyData); err != nil {
			return "", "", err
		}
		return fcert.Name(), fkey.Name(), nil
	default:
		certData, keyData, err := newCert()
		if err != nil {
			return "", "", err
		}
		fcert, err := os.CreateTemp("", "certtemp")
		if err != nil {
			return "", "", err
		}
		fkey, err := os.CreateTemp("", "keytemp")
		if err != nil {
			return "", "", err
		}
		if err := writeFile(fcert, certData); err != nil {
			return "", "", err
		}
		if err := writeFile(fkey, keyData); err != nil {
			return "", "", err
		}
		return fcert.Name(), fkey.Name(), nil
	}
}

func newCert() (*bytes.Buffer, *bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(20221),
		Subject: pkix.Name{
			Organization:  []string{"NebulaE"},
			Country:       []string{"CO"},
			Province:      []string{"Antioquia"},
			Locality:      []string{"Medellin"},
			StreetAddress: []string{"Cra. 48 #48 Sur 75"},
			PostalCode:    []string{"055422"},
			CommonName:    "localhost",
		},
		EmailAddresses:        []string{"soporte@nebulae.com.co"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(3, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return caPEM, caPrivKeyPEM, nil

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func pathExists(filename string) bool {
	dir := filepath.Dir(filename)
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isDir(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func writeFile(f *os.File, data *bytes.Buffer) error {
	defer f.Close()
	buf := make([]byte, 1024)
	for {
		if n, err := data.Read(buf); err == nil {
			if _, err = f.Write(buf[:n]); err != nil {
				return err
			}
			continue
		}
		break
	}
	return nil
}
