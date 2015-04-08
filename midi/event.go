package midi

import (
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/text/transform"
)

func writeBE(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}

func writeUvarintBE(w io.Writer, x uint32) (int, error) {
	var buf [4]byte
	n := 1
	for ; x > 0x7f; n++ {
		buf[4-n] = uint8(x & 0x7f)
		x >>= 7
	}
	buf[4-n] = uint8(x)
	for i := 4 - n; i < 3; i++ {
		buf[i] |= 0x80
	}
	return w.Write(buf[4-n:])
}

type Header struct {
	Format    uint16
	NumTracks uint16
	Division  uint16
}

func (h *Header) Size() int {
	return 14
}

func (h *Header) WriteTo(w io.Writer) (n int64, err error) {
	if err = writeBE(w, []byte("MThd")); err != nil {
		return
	}
	n += 4

	if err = writeBE(w, uint32(6)); err != nil {
		return
	}
	n += 4

	if err = writeBE(w, []uint16{
		h.Format,
		h.NumTracks,
		h.Division,
	}); err != nil {
		return
	}
	n += 6
	return
}

type TrackHeader struct {
	Len uint32
}

func (th *TrackHeader) WriteTo(w io.Writer) (n int64, err error) {
	if err = writeBE(w, []byte("MTrk")); err != nil {
		return
	}
	n += 4

	if err = writeBE(w, th.Len); err != nil {
		return
	}
	n += 4
	return
}

func (th *TrackHeader) Size() int {
	return 8
}

type DeltaTime uint32

func (d DeltaTime) Size() int {
	x, n := uint32(d), 1
	for ; x >= 0x80; n++ {
		x >>= 7
	}
	return n
}

func (d DeltaTime) WriteTo(w io.Writer) (int64, error) {
	l, err := writeUvarintBE(w, uint32(d))
	return int64(l), err
}

type NoteOnEvent struct {
	Channel  uint8
	Key      uint8
	Velocity uint8
}

func (no *NoteOnEvent) Size() int {
	return 3
}

func (no *NoteOnEvent) WriteTo(w io.Writer) (int64, error) {
	if err := writeBE(w, []byte{
		0x90 | (no.Channel & 0x0f),
		no.Key & 0x7f,
		no.Velocity & 0x7f,
	}); err != nil {
		return 0, err
	}
	return 3, nil
}

type NoteOffEvent struct {
	Channel  uint8
	Key      uint8
	Velocity uint8
}

func (no *NoteOffEvent) Size() int {
	return 3
}

func (no *NoteOffEvent) WriteTo(w io.Writer) (int64, error) {
	if err := writeBE(w, []byte{
		0x80 | (no.Channel & 0x0f),
		no.Key & 0x7f,
		no.Velocity & 0x7f,
	}); err != nil {
		return 0, err
	}
	return 3, nil
}

type TempoEvent struct {
	BPM float64
}

func (te *TempoEvent) Size() int {
	return 6
}

func (te *TempoEvent) WriteTo(w io.Writer) (int64, error) {
	bpm := uint32(60e6 / te.BPM)
	if err := writeBE(w, []byte{
		0xff,
		0x51,
		0x03,
		uint8((bpm >> 16) & 0xff),
		uint8((bpm >> 8) & 0xff),
		uint8(bpm & 0xff),
	}); err != nil {
		return 0, err
	}
	return 6, nil
}

type TimeSignatureEvent struct {
	Numerator   uint8
	Denominator uint8
}

func (te *TimeSignatureEvent) Size() int {
	return 7
}

func (te *TimeSignatureEvent) WriteTo(w io.Writer) (int64, error) {
	var d uint8 // = uint8(math.Log2(te.Denominator))
	switch te.Denominator {
	case 1:
		d = 0
	case 2:
		d = 1
	case 4:
		d = 2
	case 8:
		d = 3
	case 16:
		d = 4
	case 32:
		d = 5
	case 64:
		d = 6
	case 128:
		d = 7
	default:
		return 0, errors.New("unsupported denominator of the time signature")
	}
	if err := writeBE(w, []byte{
		0xff,
		0x58,
		0x04,
		te.Numerator,
		d,
		0x18,
		0x08,
	}); err != nil {
		return 0, err
	}
	return 7, nil
}

type TextEvent struct {
	Type        TextEventType
	Text        string
	Transformer transform.Transformer
}
type TextEventType uint8

const (
	TextEventTypeLyric = TextEventType(0x05)
)

func (te *TextEvent) Size() int {
	if te.Transformer != nil {
		buf, _, _ := transform.Bytes(te.Transformer, []byte(te.Text))
		return 3 + len(buf)
	}
	return 3 + len(te.Text)
}

func (te *TextEvent) WriteTo(w io.Writer) (n int64, err error) {
	var buf []byte
	if te.Transformer != nil {
		buf, _, err = transform.Bytes(te.Transformer, []byte(te.Text))
		if err != nil {
			return
		}
	} else {
		buf = []byte(te.Text)
	}

	if err = writeBE(w, []byte{
		0xff,
		uint8(te.Type),
		uint8(len(buf)),
	}); err != nil {
		return
	}
	n += 3

	if err = writeBE(w, buf); err != nil {
		return
	}
	n += int64(len(buf))
	return
}

type EndOfTrackEvent struct {
}

func (eote *EndOfTrackEvent) Size() int {
	return 3
}

func (eote *EndOfTrackEvent) WriteTo(w io.Writer) (int64, error) {
	if err := writeBE(w, []byte{
		0xff,
		0x2f,
		0x00,
	}); err != nil {
		return 0, err
	}
	return 3, nil
}
