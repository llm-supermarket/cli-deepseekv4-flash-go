package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/rfjakob/eme"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

const (
	fileMagic      = "RCLONE\x00\x00"
	fileMagicSize  = len(fileMagic)
	fileNonceSize  = 24
	fileHeaderSize = fileMagicSize + fileNonceSize
	blockDataSize  = 64 * 1024
	blockOverhead  = secretbox.Overhead
	blockSize      = blockOverhead + blockDataSize
	nameBlockSize  = aes.BlockSize
)

var defaultSalt = []byte{0xA8, 0x0D, 0xF4, 0x3A, 0x8F, 0xBD, 0x03, 0x08, 0xA7, 0xCA, 0xB8, 0x3E, 0x58, 0x1F, 0x86, 0xB1}

type Cipher struct {
	dataKey   [32]byte
	nameKey   [32]byte
	nameTweak [16]byte
	block     cipher.Block
	enc       NameEncoding
}

type NameEncoding int

const (
	EncodingBase32 NameEncoding = iota
	EncodingBase64
)

func ParseEncoding(s string) (NameEncoding, error) {
	switch strings.ToLower(s) {
	case "base32":
		return EncodingBase32, nil
	case "base64":
		return EncodingBase64, nil
	default:
		return EncodingBase32, fmt.Errorf("unknown filename encoding: %q", s)
	}
}

func NewCipher(password, salt string, enc NameEncoding) (*Cipher, error) {
	c := &Cipher{enc: enc}
	if err := c.deriveKeys(password, salt); err != nil {
		return nil, err
	}
	var err error
	c.block, err = aes.NewCipher(c.nameKey[:])
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cipher) deriveKeys(password, salt string) error {
	const keySize = 32 + 32 + 16
	saltBytes := defaultSalt
	if salt != "" {
		saltBytes = []byte(salt)
	}
	key, err := scrypt.Key([]byte(password), saltBytes, 16384, 8, 1, keySize)
	if err != nil {
		return err
	}
	copy(c.dataKey[:], key[:32])
	copy(c.nameKey[:], key[32:64])
	copy(c.nameTweak[:], key[64:80])
	return nil
}

func (c *Cipher) EncryptFile(in io.Reader, out io.Writer) error {
	var nonce [fileNonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	if _, err := out.Write([]byte(fileMagic)); err != nil {
		return err
	}
	if _, err := out.Write(nonce[:]); err != nil {
		return err
	}
	buf := make([]byte, blockDataSize)
	for {
		n, err := io.ReadFull(in, buf)
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			if n == 0 {
				break
			}
		} else if err != nil {
			return err
		}
		enc := secretbox.Seal(nil, buf[:n], &nonce, &c.dataKey)
		if _, err := out.Write(enc); err != nil {
			return err
		}
		incrementNonce(&nonce)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
	}
	return nil
}

func (c *Cipher) DecryptFile(in io.Reader, out io.Writer) error {
	magic := make([]byte, fileMagicSize)
	if _, err := io.ReadFull(in, magic); err != nil {
		return fmt.Errorf("failed to read magic: %w", err)
	}
	if string(magic) != fileMagic {
		return fmt.Errorf("not an rclone encrypted file - bad magic")
	}
	var nonce [fileNonceSize]byte
	if _, err := io.ReadFull(in, nonce[:]); err != nil {
		return fmt.Errorf("failed to read nonce: %w", err)
	}
	buf := make([]byte, blockSize)
	for {
		n, err := io.ReadFull(in, buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			if n == 0 {
				break
			}
		} else if err != nil {
			return err
		}
		dec, ok := secretbox.Open(nil, buf[:n], &nonce, &c.dataKey)
		if !ok {
			return fmt.Errorf("failed to decrypt block - bad password?")
		}
		if _, err := out.Write(dec); err != nil {
			return err
		}
		incrementNonce(&nonce)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
	}
	return nil
}

func (c *Cipher) EncryptFileName(plaintext string) string {
	if plaintext == "" {
		return ""
	}
	padded := pkcs7Pad(nameBlockSize, []byte(plaintext))
	ciphertext := eme.Transform(c.block, c.nameTweak[:], padded, eme.DirectionEncrypt)
	return c.encodeName(ciphertext)
}

func (c *Cipher) DecryptFileName(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	raw, err := c.decodeName(ciphertext)
	if err != nil {
		return "", err
	}
	if len(raw)%nameBlockSize != 0 {
		return "", fmt.Errorf("not a multiple of blocksize")
	}
	padded := eme.Transform(c.block, c.nameTweak[:], raw, eme.DirectionDecrypt)
	plaintext, err := pkcs7Unpad(nameBlockSize, padded)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (c *Cipher) encodeName(src []byte) string {
	switch c.enc {
	case EncodingBase32:
		encoded := base32.HexEncoding.EncodeToString(src)
		encoded = strings.TrimRight(encoded, "=")
		return strings.ToLower(encoded)
	case EncodingBase64:
		return base64.RawURLEncoding.EncodeToString(src)
	default:
		panic("unknown encoding")
	}
}

func (c *Cipher) decodeName(s string) ([]byte, error) {
	switch c.enc {
	case EncodingBase32:
		if strings.HasSuffix(s, "=") {
			return nil, fmt.Errorf("bad base32 filename encoding")
		}
		roundUp := (len(s) + 7) &^ 7
		equals := roundUp - len(s)
		s = strings.ToUpper(s) + "========"[:equals]
		return base32.HexEncoding.DecodeString(s)
	case EncodingBase64:
		return base64.RawURLEncoding.DecodeString(s)
	default:
		panic("unknown encoding")
	}
}

func incrementNonce(n *[fileNonceSize]byte) {
	for i := 0; i < fileNonceSize; i++ {
		n[i]++
		if n[i] != 0 {
			break
		}
	}
}

func pkcs7Pad(blockSize int, buf []byte) []byte {
	padding := blockSize - (len(buf) % blockSize)
	for range padding {
		buf = append(buf, byte(padding))
	}
	return buf
}

func pkcs7Unpad(blockSize int, buf []byte) ([]byte, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty buffer")
	}
	if len(buf)%blockSize != 0 {
		return nil, fmt.Errorf("not a multiple of blocksize")
	}
	padding := int(buf[len(buf)-1])
	if padding <= 0 || padding > blockSize {
		return nil, fmt.Errorf("bad padding")
	}
	for i := 0; i < padding; i++ {
		if buf[len(buf)-1-i] != byte(padding) {
			return nil, fmt.Errorf("bad padding")
		}
	}
	return buf[:len(buf)-padding], nil
}
