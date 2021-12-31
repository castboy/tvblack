package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type PriceEncrypt struct {
	enc_key string
	sig_key string

	enc_key_bytes []byte
	sig_key_bytes []byte
}

func NewPriceEncrypt(enc_key, sig_key string) *PriceEncrypt {
	return &PriceEncrypt{enc_key: enc_key, sig_key: sig_key,
		enc_key_bytes: hex2bin(enc_key), sig_key_bytes: hex2bin(sig_key)}
}

func (this *PriceEncrypt) Encode(price float64) string {
	price_buf := new(bytes.Buffer)
	binary.Write(price_buf, binary.LittleEndian, price)

	time_ms_stamp := time.Now().Unix()
	time_ms_stamp_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(time_ms_stamp_bytes, uint32(time_ms_stamp))

	pad := sha1_hmac(time_ms_stamp_bytes, this.enc_key_bytes)
	enc_price := xor_bytes(pad, price_buf.Bytes(), 8)

	sig := sha1_hmac(append(price_buf.Bytes(), time_ms_stamp_bytes...), this.sig_key_bytes)
	sig = sig[:4]
	ret := base64url_encode(append(append(time_ms_stamp_bytes, enc_price...), sig...))

	return strings.TrimRight(ret, "=")
}

func (this *PriceEncrypt) Decode(base64Str string) (float64, error) {
	data, err := base64url_decode(base64Str)
	if err != nil {
		return 0, err
	}

	if len(data) != 16 {
		return 0, errors.New("Illegal base64 string")
	}

	time_ms_stamp_bytes := data[:4]
	enc_price := data[4:12]
	sig := data[12:16]

	pad := sha1_hmac(time_ms_stamp_bytes, this.enc_key_bytes)
	dec_price := xor_bytes(pad, enc_price, 8)

	var price float64

	dec_price_buf := bytes.NewBuffer(dec_price)
	binary.Read(dec_price_buf, binary.LittleEndian, &price)

	osig := sha1_hmac(append(dec_price, time_ms_stamp_bytes...), this.sig_key_bytes)
	if bytes.Compare(sig, osig[:len(sig)]) == 0 {
		return price, nil
	}

	return price, errors.New("signature is illegal")
}

func xor_bytes(b1, b2 []byte, length int) []byte {
	new_b := make([]byte, length)
	for i := 0; i < length; i++ {
		new_b[i] = b1[i] ^ b2[i]
	}
	return new_b
}

func hex2bin(s string) []byte {
	ret, _ := hex.DecodeString(s)
	return ret
}

func sha1_hmac(data, key []byte) []byte {
	mac := hmac.New(sha1.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func base64url_encode(data []byte) string {
	ret := base64.StdEncoding.EncodeToString(data)
	return strings.Map(func(r rune) rune {
		switch r {
		case '+':
			return '-'
		case '/':
			return '_'
		}

		return r
	}, ret)
}

func base64url_decode(s string) ([]byte, error) {
	base64Str := strings.Map(func(r rune) rune {
		switch r {
		case '-':
			return '+'
		case '_':
			return '/'
		}

		return r
	}, s)

	if pad := len(base64Str) % 4; pad > 0 {
		base64Str += strings.Repeat("=", 4-pad)
	}

	return base64.StdEncoding.DecodeString(base64Str)
}
