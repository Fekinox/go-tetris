package main

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
)

type ReplayData struct {
	Seed              int64
	TetrisSettings    GlobalTetrisSettings
	ObjectiveID       ObjectiveID
	ObjectiveSettings ObjectiveSettings
	Actions           []ReplayAction
}

func (rd *ReplayData) Encode(w io.Writer) error {
	base64Encoder := base64.NewEncoder(base64.StdEncoding, w)
	err := binary.Write(base64Encoder, binary.LittleEndian, rd.Seed)
	if err != nil {
		return err
	}
	err = binary.Write(base64Encoder, binary.LittleEndian, rd.TetrisSettings)
	if err != nil {
		return err
	}
	err = binary.Write(base64Encoder, binary.LittleEndian, rd.ObjectiveID)
	if err != nil {
		return err
	}
	switch set := rd.ObjectiveSettings.(type) {
	case *LineClearSettings:
		err = binary.Write(base64Encoder, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	case *SurvivalSettings:
		err = binary.Write(base64Encoder, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	case *EndlessSettings:
		err = binary.Write(base64Encoder, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	case *CheeseSettings:
		err = binary.Write(base64Encoder, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	}
	err = binary.Write(base64Encoder, binary.LittleEndian, int64(len(rd.Actions)))
	if err != nil {
		return err
	}

	for i := 0; i < len(rd.Actions); i++ {
		err = binary.Write(base64Encoder, binary.LittleEndian, rd.Actions[i])
		if err != nil {
			return err
		}
	}
	return base64Encoder.Close()
}

func (rd *ReplayData) Decode(r io.Reader) error {
	base64Decoder := base64.NewDecoder(base64.StdEncoding, r)

	err := binary.Read(base64Decoder, binary.LittleEndian, &rd.Seed)
	if err != nil {
		return err
	}

	err = binary.Read(base64Decoder, binary.LittleEndian, &rd.TetrisSettings)
	if err != nil {
		return err
	}

	err = binary.Read(base64Decoder, binary.LittleEndian, &rd.ObjectiveID)
	if err != nil {
		return err
	}

	switch rd.ObjectiveID {
	case LineClear:
		var lineclear LineClearSettings
		err = binary.Read(base64Decoder, binary.LittleEndian, &lineclear)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &lineclear
	case Survival:
		var survival SurvivalSettings
		err = binary.Read(base64Decoder, binary.LittleEndian, &survival)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &survival
	case Endless:
		var endless EndlessSettings
		err = binary.Read(base64Decoder, binary.LittleEndian, &endless)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &endless
	case Cheese:
		var cheese CheeseSettings
		err = binary.Read(base64Decoder, binary.LittleEndian, &cheese)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &cheese
	default:
		return errors.New("Invalid objective ID")
	}

	var numActions int64
	err = binary.Read(base64Decoder, binary.LittleEndian, &numActions)
	if err != nil {
		return err
	}

	rd.Actions = make([]ReplayAction, numActions)
	var i int64
	for i = 0; i < numActions; i++ {
		err = binary.Read(base64Decoder, binary.LittleEndian, &rd.Actions[i])
		if err != nil {
			return err
		}
	}

	return nil
}
