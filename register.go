package jwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

// KeyRegister contains recognized credentials.
type KeyRegister struct {
	ECDSAs  []*ecdsa.PublicKey // ECDSA credentials
	RSAs    []*rsa.PublicKey   // RSA credentials
	Secrets [][]byte           // HMAC credentials
}

// Check parses a JWT and returns the claims set if, and only if, the signature
// checks out. Note that this excludes unsecured JWTs [ErrUnsecured].
// See Claims.Valid to complete the verification.
func (reg *KeyRegister) Check(token []byte) (*Claims, error) {
	header, buf, err := parseHeader(token)
	if err != nil {
		return nil, err
	}

	var verifySig func(content, sig []byte, hash crypto.Hash) error
	hash, err := header.match(HMACAlgs)
	if err == nil {
		verifySig = func(content, sig []byte, hash crypto.Hash) error {
			for _, secret := range reg.Secrets {
				digest := hmac.New(hash.New, secret)
				digest.Write(content)
				if hmac.Equal(sig, digest.Sum(sig[len(sig):])) {
					return nil
				}
			}
			return ErrSigMiss
		}
	} else if err != ErrAlgUnk {
		return nil, err
	} else if hash, err = header.match(RSAAlgs); err == nil {
		verifySig = func(content, sig []byte, hash crypto.Hash) error {
			digest := hash.New()
			digest.Write(content)
			digestSum := digest.Sum(sig[len(sig):])
			for _, key := range reg.RSAs {
				if err := rsa.VerifyPKCS1v15(key, hash, digestSum, sig); err == nil {
					return nil
				}
			}
			return ErrSigMiss
		}
	} else if err != ErrAlgUnk {
		return nil, err
	} else if hash, err = header.match(ECDSAAlgs); err == nil {
		verifySig = func(content, sig []byte, hash crypto.Hash) error {
			r := big.NewInt(0).SetBytes(sig[:len(sig)/2])
			s := big.NewInt(0).SetBytes(sig[len(sig)/2:])
			digest := hash.New()
			digest.Write(content)
			digestSum := digest.Sum(sig[:0])
			for _, key := range reg.ECDSAs {
				if ecdsa.Verify(key, digestSum, r, s) {
					return nil
				}
			}
			return ErrSigMiss
		}
	} else {
		return nil, err
	}

	claims, err := verifyAndParseClaims(token, buf, hash, verifySig)
	if err != nil {
		return nil, err
	}

	claims.KeyID = header.Kid
	return claims, nil
}

var errUnencryptedPEM = errors.New("jwt: unencrypted PEM rejected due password expectation")

// LoadPEM adds keys from PEM-encoded data and returns the count. PEM encryption
// is enforced for non-empty password values. The source may be certificates,
// public keys, private keys, or a combination of any of the previous. Private
// keys are discared after the (automatic) public key extraction completes.
func (r *KeyRegister) LoadPEM(data, password []byte) (n int, err error) {
	for {
		block, remainder := pem.Decode(data)
		if block == nil {
			return
		}
		data = remainder

		if x509.IsEncryptedPEMBlock(block) {
			block.Bytes, err = x509.DecryptPEMBlock(block, password)
			if err != nil {
				return
			}
		} else if len(password) != 0 {
			return n, errUnencryptedPEM
		}

		switch block.Type {
		case "CERTIFICATE":
			certs, err := x509.ParseCertificates(block.Bytes)
			if err != nil {
				return n, err
			}
			for _, c := range certs {
				if err := r.add(c.PublicKey); err != nil {
					return n, err
				}
			}

		case "PUBLIC KEY":
			key, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return n, err
			}
			if err := r.add(key); err != nil {
				return n, err
			}

		case "EC PRIVATE KEY":
			key, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return n, err
			}
			r.ECDSAs = append(r.ECDSAs, &key.PublicKey)

		case "RSA PRIVATE KEY":
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return n, err
			}
			r.RSAs = append(r.RSAs, &key.PublicKey)

		default:
			return n, fmt.Errorf("jwt: unknown PEM type %q", block.Type)
		}

		n++
	}
}

func (r *KeyRegister) add(key interface{}) error {
	switch t := key.(type) {
	case *ecdsa.PublicKey:
		r.ECDSAs = append(r.ECDSAs, t)
	case *rsa.PublicKey:
		r.RSAs = append(r.RSAs, t)
	default:
		return fmt.Errorf("jwt: unsupported key type %T", t)
	}
	return nil
}
