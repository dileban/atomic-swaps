package security

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
)

// X509Certificate embeds an x509.Certificate and implements the
// Identity interface.
type X509Certificate struct {
	*x509.Certificate
}

// X509Identity interface allows for accessing properties of the
// underlying x509 identity representation. The current interface
// allows retrieving various forms of addresses compatible with
// different blockchain protocol implementations.
type X509Identity interface {

	// Returns the public key associated with the underlying identity.
	GetPublicKey() interface{}

	// GetAddress returns a custom string representation of the public
	// key.
	GetAddress() string

	// GetBitcoinAddress returns a Bitcoin compatiable address based on
	// the public key.
	GetBitcoinAddress() string

	// GetEthereumAddress returns an Ethereum compatible address based
	// on the public key.
	GetEthereumAddress() string

	// GetAttribute returns the attribute value of the specified key.
	GetAttribute(key string) string

	// GetIssuer returns the issuing authority of the x509 certificate.
	GetIssuer() string
}

// NewX509Certificate extends an x509.Certificate instance with
// a set of convenience methods.
func NewX509Certificate(cert *x509.Certificate) *X509Certificate {
	return &X509Certificate{cert}
}

// GetAddress returns a 64 character hex representation of the public
// key.
func (c *X509Certificate) GetAddress() string {
	pub := publicKeyToBytes(c.PublicKey)
	shaPub := sha256.Sum256(pub)
	return hex.EncodeToString(shaPub[:])
}

// GetBitcoinAddress returns a Bitcoin compatiable address based on
// the public key.
func (c *X509Certificate) GetBitcoinAddress() string {
	// TODO: Implement GetBitcoinAddress
	return ""
}

// GetEthereumAddress returns an Ethereum compatible address based
// on the public key.
func (c *X509Certificate) GetEthereumAddress() string {
	// TODO: Implement GetEthereumAddress
	return ""
}

// publicKeyToBytes converts a public key based on one of RSA, DSA or
// ECDSA to a byte array.
func publicKeyToBytes(pub interface{}) []byte {
	var b []byte
	switch k := pub.(type) {
	case *rsa.PublicKey:
		b = k.N.Bytes()
	case *dsa.PublicKey:
		b = k.Y.Bytes()
	case *ecdsa.PublicKey:
		b = append(k.X.Bytes(), k.Y.Bytes()...)
	}
	return b
}
