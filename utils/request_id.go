package utils

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"time"

	"github.com/nihuo/go-base62"
)

const (
	padPrefixLen  = 6
	padTimeLen    = 6
	padRandNumLen = 6

	randNumCount     = 3
	randNumBytesStep = 4
)

type RequestID struct {
	prefixStr string
}

func NewRequestID(prefixStr []byte) (*RequestID, error) {
	prefixNum := binary.BigEndian.Uint32(prefixStr)
	s := base62.Encode(uint(prefixNum))
	return &RequestID{
		prefixStr: leftPad2Len(s, "0", padPrefixLen),
	}, nil
}

func (r *RequestID) GenerateID() string {
	timeStr := leftPad2Len(base62.Encode(uint(time.Now().Unix())), "0", padTimeLen)
	return r.prefixStr + timeStr + r.randomNumStr()
}

func (r *RequestID) DecryptID(id string) ([]byte, int, error) {
	if len(id) != (padPrefixLen + padTimeLen + padRandNumLen*randNumCount) {
		return []byte{}, 0, errors.New("request id length Invalid")
	}

	prefixByte := make([]byte, 4)
	prefixNum := base62.Decode(id[:padPrefixLen])
	binary.BigEndian.PutUint32(prefixByte, uint32(prefixNum))

	return prefixByte, int(base62.Decode(id[padPrefixLen : padPrefixLen+padTimeLen])), nil
}

func (r *RequestID) randomNumStr() string {
	randBytes := make([]byte, randNumBytesStep*randNumCount)
	_, err := rand.Read(randBytes)
	if err != nil {
		return ""
	}

	numStr := ""
	for i := 0; i < randNumCount; i++ {
		n := binary.LittleEndian.Uint32(randBytes[i*randNumBytesStep : (i+1)*randNumBytesStep])

		s := base62.Encode(uint(n))
		numStr += leftPad2Len(s, "0", padRandNumLen)
	}

	return numStr
}
