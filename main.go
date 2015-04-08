package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"

	"github.com/oov/mxl2mid/mxl"
)

func readMXML(filename string) (*mxl.MXLDoc, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return mxl.Decode(f)
}

func writeMIDI(filename string, mxldoc *mxl.MXLDoc, metaTextTransformer transform.Transformer) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = mxldoc.MIDI(metaTextTransformer).WriteTo(f)
	return err
}

func main() {
	cs := flag.String("charset", "Shift_JIS", "meta info charset of the output destination midi file")
	flag.Parse()

	e, _ := charset.Lookup(*cs)
	lyricTransformer := e.NewEncoder()

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
	err = writeMIDI(outfile, mxl, lyricTransformer)
	if err != nil {
		log.Fatal(err)
	}
}
