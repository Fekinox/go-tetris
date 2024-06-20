package main

import (
	"encoding/gob"
	"encoding/base64"
	"io"
)

type ReplayData struct {
	Seed	int64
	TetrisSettings GlobalTetrisSettings
	ObjectiveSettings ObjectiveSettings
	Actions []ReplayAction
}

func (rd *ReplayData) Encode(w io.Writer) error {
	gob.Register(rd.ObjectiveSettings)

	base64Encoder := base64.NewEncoder(base64.StdEncoding, w)
	gobEncoder := gob.NewEncoder(base64Encoder)
	err := gobEncoder.Encode(rd)
	if err != nil {
		return err
	}
	return base64Encoder.Close()
}

func (rd *ReplayData) Decode(r io.Reader) error {
	base64Decoder := base64.NewDecoder(base64.StdEncoding, r)
	gobDecoder := gob.NewDecoder(base64Decoder)
	return gobDecoder.Decode(rd)
}
