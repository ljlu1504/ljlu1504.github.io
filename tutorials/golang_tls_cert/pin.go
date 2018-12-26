package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/runner"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

// KeyPairWithPin returns PEM encoded Certificate and Key along with an SKPI
// fingerprint of the public key.
func KeyPairWithPin() ([]byte, []byte, []byte, error) {
	bits := 4096
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "rsa.GenerateKey")
	}

	tpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "169.264.169.254"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	derCert, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "x509.CreateCertificate")
	}

	buf := &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCert,
	})
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "pem.Encode")
	}

	pemCert := buf.Bytes()

	buf = &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "pem.Encode")
	}
	pemKey := buf.Bytes()
	cert, err := x509.ParseCertificate(derCert)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "x509.ParseCertificate")
	}

	pubDER, err := x509.MarshalPKIXPublicKey(cert.PublicKey.(*rsa.PublicKey))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "x509.MarshalPKIXPublicKey")
	}
	sum := sha256.Sum256(pubDER)
	pin := make([]byte, base64.StdEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(pin, sum[:])

	return pemCert, pemKey, pin, nil
}

func main()  {
	cdp, err := chromedp.New(
		ctx,
		chromedp.WithRunnerOptions(
			// ...
			runner.Flag("ignore-certificate-errors-spki-list", pin),
		),
	)
}
