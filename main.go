package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"

	"github.com/oov/mxl2mid/midi"
	"github.com/oov/mxl2mid/mxl"
)

var lyricTransformer transform.Transformer

func readMXML(filename string) (*mxl.MXLDoc, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	dec.CharsetReader = charset.NewReaderLabel

	var mxl mxl.MXLDoc
	err = dec.Decode(&mxl)
	if err != nil {
		return nil, err
	}

	return &mxl, nil
}

func writeMIDI(filename string, mxldoc *mxl.MXLDoc) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = (&midi.Header{
		Format:    1,
		NumTracks: 2,
		Division:  uint16(mxldoc.FindDivisions()),
	}).WriteTo(f)
	if err != nil {
		return err
	}

	// build conductor track

	// Mtrk placeholder
	var trackHeader midi.TrackHeader
	_, err = trackHeader.WriteTo(f)
	if err != nil {
		return err
	}

	var d int
	var written, l int64
	for _, measure := range mxldoc.Parts[0].Measures {
		if measure.Attrs.Time.Beats != 0 {
			l, err = (&midi.TimeSignatureEvent{
				DeltaTime:   midi.DeltaTime(d),
				Numerator:   uint8(measure.Attrs.Time.Beats),
				Denominator: uint8(measure.Attrs.Time.BeatType),
			}).WriteTo(f)
			if err != nil {
				return err
			}
			written += l
			d = 0
		}
		for _, event := range measure.Events {
			switch v := event.(type) {
			case mxl.Sound:
				l, err = (&midi.TempoEvent{
					DeltaTime: midi.DeltaTime(d),
					BPM:       v.Tempo,
				}).WriteTo(f)
				if err != nil {
					return err
				}
				written += l
				d = 0
			case mxl.Note:
				d += v.Duration
			}
		}
	}

	l, err = (&midi.EndOfTrackEvent{
		DeltaTime: midi.DeltaTime(d),
	}).WriteTo(f)
	if err != nil {
		return err
	}
	written += l

	trackHeader.Len = uint32(written)
	_, err = f.Seek(-written-int64(trackHeader.Size()), os.SEEK_CUR)
	if err != nil {
		return err
	}

	_, err = trackHeader.WriteTo(f)
	if err != nil {
		return err
	}

	_, err = f.Seek(written, os.SEEK_CUR)
	if err != nil {
		return err
	}

	// build main track

	// Mtrk placeholder
	trackHeader.Len = 0
	_, err = trackHeader.WriteTo(f)
	if err != nil {
		return err
	}

	d, written, l = 0, 0, 0
	for _, measure := range mxldoc.Parts[0].Measures {
		for _, event := range measure.Events {
			switch v := event.(type) {
			case mxl.Note:
				switch {
				case v.Rest.Local != "":
					d += v.Duration
				case v.Pitch.Step != "":
					if v.Tie.Type != "stop" {
						// lyric
						l, err = (&midi.TextEvent{
							DeltaTime:   midi.DeltaTime(d),
							Type:        midi.TextEventTypeLyric,
							Text:        v.Lyric.Text,
							Transformer: lyricTransformer,
						}).WriteTo(f)
						if err != nil {
							return err
						}
						written += l
						d = 0

						// note on
						l, err = (&midi.NoteOnEvent{
							DeltaTime: midi.DeltaTime(0),
							Channel:   0,
							Key:       uint8(v.Pitch.Key()),
							Velocity:  100,
						}).WriteTo(f)
						if err != nil {
							return err
						}
						written += l
					}

					d += v.Duration
					if v.Tie.Type != "start" {
						// note off
						l, err = (&midi.NoteOffEvent{
							DeltaTime: midi.DeltaTime(d),
							Channel:   0,
							Key:       uint8(v.Pitch.Key()),
							Velocity:  0,
						}).WriteTo(f)
						if err != nil {
							return err
						}
						written += l
						d = 0
					}
				}
			}
		}
	}

	l, err = (&midi.EndOfTrackEvent{
		DeltaTime: midi.DeltaTime(d),
	}).WriteTo(f)
	if err != nil {
		return err
	}
	written += l

	trackHeader.Len = uint32(written)
	_, err = f.Seek(-written-int64(trackHeader.Size()), os.SEEK_CUR)
	if err != nil {
		return err
	}

	_, err = trackHeader.WriteTo(f)
	if err != nil {
		return err
	}

	_, err = f.Seek(written, os.SEEK_CUR)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	cs := flag.String("charset", "Shift_JIS", "meta info charset of the output destination midi file")
	flag.Parse()

	e, _ := charset.Lookup(*cs)
	lyricTransformer = e.NewEncoder()

	if flag.NArg() == 0 {
		fmt.Println(`mxl2mid
=======

This is a MusicXML to MIDI converter.
This program supports only MusicXML file that was output from CeVIO Creative Studio.

mxl2mid [-charset=Shift_JIS] infile [outfile]
`)
		flag.Usage()
		return
	}
	infile := flag.Arg(0)
	mxl, err := readMXML(infile)
	if err != nil {
		log.Fatal(err)
	}

	var outfile string
	if flag.NArg() >= 2 {
		outfile = flag.Arg(1)
	} else {
		outfile = infile + ".mid"
	}
	err = writeMIDI(outfile, mxl)
	if err != nil {
		log.Fatal(err)
	}
}
