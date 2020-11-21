package lib

import (
	"bytes"
	"encoding/hex"
)

type fileData []byte

func (fdat fileData) status() Status {
	return Status(fdat[0])
}

func (fdat fileData) setStatus(status Status) {
	fdat[0] = byte(status)
}

func (fdat fileData) checksum() []byte {
	return fdat[1:]
}

func (fdat fileData) checksumString() string {
	return hex.EncodeToString(fdat.checksum())
}

func (fdat fileData) checksumEqual(checksum []byte) bool {
	return bytes.Equal(fdat.checksum(), checksum)
}

func makeFileData(checksum []byte, status Status) (fdat fileData) {
	fdat = make([]byte, len(checksum)+1)
	fdat.setStatus(status)
	copy(fdat[1:], checksum)
	return
}
