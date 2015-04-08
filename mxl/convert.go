package mxl

import (
	"golang.org/x/text/transform"

	"github.com/oov/mxl2mid/midi"
)

type builder struct {
	mxldoc              *MXLDoc
	metaTextTransformer transform.Transformer
}

func (br *builder) buildConductorTrack() midi.Track {
	var tr midi.TrackBuilder
	for _, measure := range br.mxldoc.Parts[0].Measures {
		if measure.Attrs.Time.Beats != 0 {
			tr.AddEvent(&midi.TimeSignatureEvent{
				Numerator:   uint8(measure.Attrs.Time.Beats),
				Denominator: uint8(measure.Attrs.Time.BeatType),
			})
		}
		for _, event := range measure.Events {
			switch v := event.(type) {
			case Sound:
				tr.AddEvent(&midi.TempoEvent{BPM: v.Tempo})
			case Note:
				tr.AddDeltaTime(v.Duration)
			}
		}
	}

	tr.AddEvent(&midi.EndOfTrackEvent{})
	return tr.Track
}

func (br *builder) buildMainTrack() midi.Track {
	var tr midi.TrackBuilder
	for _, measure := range br.mxldoc.Parts[0].Measures {
		for _, event := range measure.Events {
			switch v := event.(type) {
			case Note:
				switch {
				case v.Rest.Local != "":
					tr.AddDeltaTime(v.Duration)
				case v.Pitch.Step != "":
					if v.Tie.Type != "stop" {
						tr.AddEvent(&midi.TextEvent{
							Type:        midi.TextEventTypeLyric,
							Text:        v.Lyric.Text,
							Transformer: br.metaTextTransformer,
						})

						tr.AddEvent(&midi.NoteOnEvent{
							Channel:  0,
							Key:      uint8(v.Pitch.Key()),
							Velocity: 100,
						})
					}

					tr.AddDeltaTime(v.Duration)
					if v.Tie.Type != "start" {
						tr.AddEvent(&midi.NoteOffEvent{
							Channel:  0,
							Key:      uint8(v.Pitch.Key()),
							Velocity: 0,
						})
					}
				}
			}
		}
	}

	tr.AddEvent(&midi.EndOfTrackEvent{})
	return tr.Track
}

func (d *MXLDoc) MIDI(metaTextTransformer transform.Transformer) *midi.MIDI {
	br := builder{
		mxldoc:              d,
		metaTextTransformer: metaTextTransformer,
	}
	return &midi.MIDI{
		Format:   1,
		Division: uint16(d.FindDivisions()),
		Tracks: []midi.Track{
			br.buildConductorTrack(),
			br.buildMainTrack(),
		},
	}
}
