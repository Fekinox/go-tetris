package main

import (
	"encoding/gob"
	"encoding/base64"
	"io"
)

type ReplayData struct {
	Seed	int64
	TetrisSettings GlobalTetrisSettings
	ObjectiveID ObjectiveID
	ObjectiveSettings ObjectiveSettings
	Actions []ReplayAction
}

func (rd *ReplayData) Encode(w io.Writer) error {
	base64Encoder := base64.NewEncoder(base64.StdEncoding, w)
	gobEncoder := gob.NewEncoder(base64Encoder)
	return gobEncoder.Encode(rd)
}

func (rd *ReplayData) Decode(r io.Reader) error {
	base64Decoder := base64.NewDecoder(base64.StdEncoding, r)
	gobDecoder := gob.NewDecoder(base64Decoder)
	return gobDecoder.Decode(rd)
}
