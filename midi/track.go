package midi

import (
	"io"
)

type Event interface {
	io.WriterTo
	Size() int
}

type DeltaTimeEvent struct {
	DeltaTime
	Event Event
}

type Track []DeltaTimeEvent

func (t Track) WriteTo(w io.Writer) (int64, error) {
	ln := 0
	for _, de := range t {
		ln += de.DeltaTime.Size() + de.Event.Size()
	}

	written, err := (&TrackHeader{Len: uint32(ln)}).WriteTo(w)
	if err != nil {
		return 0, err
	}

	var l int64
	for _, de := range t {
		l, err = de.DeltaTime.WriteTo(w)
		if err != nil {
			return written, err
		}
		written += l

		l, err = de.Event.WriteTo(w)
		if err != nil {
			return written, err
		}
		written += l
	}

	return written, nil
}

type TrackBuilder struct {
	d     int
	Track []DeltaTimeEvent
}

func (tb *TrackBuilder) AddEvent(event Event) {
	tb.Track = append(tb.Track, DeltaTimeEvent{
		DeltaTime: DeltaTime(tb.d),
		Event:     event,
	})
	tb.d = 0
}

func (tb *TrackBuilder) AddDeltaTime(d int) {
	tb.d += d
}

type MIDI struct {
	Format   uint16
	Division uint16
	Tracks   []Track
}

func (m *MIDI) WriteTo(w io.Writer) (int64, error) {
	written, err := (&Header{
		Format:    m.Format,
		NumTracks: uint16(len(m.Tracks)),
		Division:  m.Division,
	}).WriteTo(w)
	if err != nil {
		return 0, err
	}

	var l int64
	for _, t := range m.Tracks {
		l, err = t.WriteTo(w)
		if err != nil {
			return written, err
		}
		written += l
	}
	return written, nil
}
