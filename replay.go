package main

import (
	"compress/gzip"
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

type ReplayEncoder func(rd *ReplayData, w io.Writer) error
type ReplayDecoder func(r io.Reader) (*ReplayData, error)

var StdEncoder ReplayEncoder = EncodeCompressed
var StdDecoder ReplayDecoder = DecodeCompressed

func EncodeUncompressed(rd *ReplayData, w io.Writer) error {
	base64Encoder := base64.NewEncoder(base64.StdEncoding, w)

	err := rd.Encode(base64Encoder)
	if err != nil {
		return err
	}

	err = base64Encoder.Close()
	if err != nil {
		return err
	}

	return nil
}

func DecodeUncompressed(r io.Reader) (*ReplayData, error) {
	rd := ReplayData{}
	base64Decoder := base64.NewDecoder(base64.StdEncoding, r)

	err := rd.Decode(base64Decoder)
	if err != nil {
		return nil, err
	}

	return &rd, nil
}

func EncodeCompressed(rd *ReplayData, w io.Writer) error {
	base64Encoder := base64.NewEncoder(base64.StdEncoding, w)
	gzipEncoder := gzip.NewWriter(base64Encoder)

	err := rd.Encode(gzipEncoder)
	if err != nil {
		return err
	}

	err = gzipEncoder.Close()
	if err != nil {
		return err
	}

	err = base64Encoder.Close()
	if err != nil {
		return err
	}

	return nil
}

func DecodeCompressed(r io.Reader) (*ReplayData, error) {
	rd := ReplayData{}
	var err error
	base64Decoder := base64.NewDecoder(base64.StdEncoding, r)
	gzipDecoder, err := gzip.NewReader(base64Decoder)
	if err != nil {
		return nil, err
	}

	err = rd.Decode(gzipDecoder)
	if err != nil {
		return nil, err
	}

	err = gzipDecoder.Close()
	if err != nil {
		return nil, err
	}

	return &rd, nil
}

func (rd *ReplayData) Encode(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, rd.Seed)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, rd.TetrisSettings)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, rd.ObjectiveID)
	if err != nil {
		return err
	}
	switch set := rd.ObjectiveSettings.(type) {
	case *LineClearSettings:
		err = binary.Write(w, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	case *SurvivalSettings:
		err = binary.Write(w, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	case *EndlessSettings:
		err = binary.Write(w, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	case *CheeseSettings:
		err = binary.Write(w, binary.LittleEndian, set)
		if err != nil {
			return err
		}
	}
	err = binary.Write(
		w,
		binary.LittleEndian,
		int64(len(rd.Actions)),
	)
	if err != nil {
		return err
	}

	for i := 0; i < len(rd.Actions); i++ {
		err = binary.Write(w, binary.LittleEndian, rd.Actions[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (rd *ReplayData) Decode(r io.Reader) error {
	var err error
	err = binary.Read(r, binary.LittleEndian, &rd.Seed)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.LittleEndian, &rd.TetrisSettings)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.LittleEndian, &rd.ObjectiveID)
	if err != nil {
		return err
	}

	switch rd.ObjectiveID {
	case LineClear:
		var lineclear LineClearSettings
		err = binary.Read(r, binary.LittleEndian, &lineclear)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &lineclear
	case Survival:
		var survival SurvivalSettings
		err = binary.Read(r, binary.LittleEndian, &survival)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &survival
	case Endless:
		var endless EndlessSettings
		err = binary.Read(r, binary.LittleEndian, &endless)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &endless
	case Cheese:
		var cheese CheeseSettings
		err = binary.Read(r, binary.LittleEndian, &cheese)
		if err != nil {
			return err
		}
		rd.ObjectiveSettings = &cheese
	default:
		return errors.New("Invalid objective ID")
	}

	var numActions int64
	err = binary.Read(r, binary.LittleEndian, &numActions)
	if err != nil {
		return err
	}

	rd.Actions = make([]ReplayAction, numActions)
	var i int64
	for i = 0; i < numActions; i++ {
		err = binary.Read(r, binary.LittleEndian, &rd.Actions[i])
		if err != nil {
			return err
		}
	}

	return nil
}
