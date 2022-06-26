package tun

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

type sessionDesc struct {
	Address string `json:"address"`
	UUID    string `json:"uuid"`
}

func readDesc(r io.Reader) (*sessionDesc, error) {
	h := make([]byte, 2)
	_, err := io.ReadFull(r, h)
	if err != nil {
		return nil, err
	}
	len := binary.BigEndian.Uint16(h)
	body := make([]byte, len)
	_, err = io.ReadFull(r, body)
	if err != nil {
		return nil, err
	}
	ret := &sessionDesc{}
	if err := json.Unmarshal(body, ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func writeDesc(w io.Writer, desc sessionDesc) error {
	body, err := json.Marshal(desc)
	if err != nil {
		return err
	}
	b := make([]byte, 2+len(body))
	binary.BigEndian.PutUint16(b, uint16(len(body)))
	copy(b[2:], body)
	_, err = w.Write(b)
	return err
}
