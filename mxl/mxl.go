// this package is based on https://github.com/eliothedeman/go-mxl
//
// The MIT License (MIT)
//
// Copyright (c) 2014 Eliot Hedeman
// Copyright (c) 2015 oov
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package mxl

import (
	"encoding/xml"
	"io"
	"strconv"

	"golang.org/x/net/html/charset"
)

// MXLDoc holds all data for a music xml file
type MXLDoc struct {
	Score          xml.Name `xml:"score-partwise"`
	Identification `xml:"identification"`
	Parts          []Part `xml:"part"`
}

func (d *MXLDoc) FindDivisions() int {
	for _, part := range d.Parts {
		for _, measure := range part.Measures {
			if measure.Attrs.Divisions != 0 {
				return measure.Attrs.Divisions
			}
		}
	}
	return 0
}

// Identification holds all of the ident information for a music xml file
type Identification struct {
	Composer string   `xml:"creator"`
	Encoding Encoding `xml:"encoding"`
	Rights   string   `xml:"rights"`
	Source   string   `xml:"source"`
	Title    string   `xml:"movement-title"`
}

// Encoding holds encoding info
type Encoding struct {
	Software string `xml:"software"`
	Date     string `xml:"encoding-date"`
}

// Part represents a part in a piece of music
type Part struct {
	Id       string    `xml:"id,attr"`
	Measures []Measure `xml:"measure"`
}

// Measure represents a measure in a piece of music
type Measure struct {
	Number int        `xml:"number,attr"`
	Attrs  Attributes `xml:"attributes"`
	Events []interface{}
}

func (m *Measure) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "number" {
			m.Number, _ = strconv.Atoi(attr.Value)
		}
	}

	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if t, ok := token.(xml.StartElement); ok {
			switch t.Name.Local {
			case "attributes":
				if err := d.DecodeElement(&m.Attrs, &t); err != nil {
					return err
				}
			case "sound":
				var snd Sound
				if err := d.DecodeElement(&snd, &t); err != nil {
					return err
				}
				m.Events = append(m.Events, snd)
			case "note":
				var n Note
				if err := d.DecodeElement(&n, &t); err != nil {
					return err
				}
				m.Events = append(m.Events, n)
			case "direction":
				d.Skip()
			}
		}
	}
	return nil
}

// Attributes represents
type Attributes struct {
	Key       Key  `xml:"key"`
	Time      Time `xml:"time"`
	Divisions int  `xml:"divisions"`
	Clef      Clef `xml:"clef"`
}

type Sound struct {
	Tempo float64 `xml:"tempo,attr"`
}

// Clef represents a clef change
type Clef struct {
	Sign string `xml:"sign"`
	Line int    `xml:"line"`
}

// Key represents a key signature change
type Key struct {
	Fifths int    `xml:"fifths"`
	Mode   string `xml:"mode"`
}

// Time represents a time signature change
type Time struct {
	Beats    int `xml:"beats"`
	BeatType int `xml:"beat-type"`
}

// Note represents a note in a measure
type Note struct {
	Pitch    Pitch    `xml:"pitch"`
	Lyric    Lyric    `xml:"lyric"`
	Duration int      `xml:"duration"`
	Voice    int      `xml:"voice"`
	Type     string   `xml:"type"`
	Tie      Tie      `xml:"tie"`
	Rest     xml.Name `xml:"rest"`
	Chord    xml.Name `xml:"chord"`
}

type Tie struct {
	Type string `xml:"type,attr"`
}

// Pitch represents the pitch of a note
type Pitch struct {
	Accidental int8   `xml:"alter"`
	Step       string `xml:"step"`
	Octave     int    `xml:"octave"`
}

func (p *Pitch) Key() int {
	var n int
	switch p.Step {
	case "C":
		n = 0
	case "D":
		n = 2
	case "E":
		n = 4
	case "F":
		n = 5
	case "G":
		n = 7
	case "A":
		n = 9
	case "B":
		n = 11
	}
	return n + (p.Octave+1)*12 + int(p.Accidental)

}

// Lyric represents the lyric of a note
type Lyric struct {
	Syllabic string `xml:"syllabic"`
	Text     string `xml:"text"`
}

func Decode(r io.Reader) (*MXLDoc, error) {
	var mxl MXLDoc
	dec := xml.NewDecoder(r)
	dec.CharsetReader = charset.NewReaderLabel
	err := dec.Decode(&mxl)
	if err != nil {
		return nil, err
	}
	return &mxl, nil
}
