package main

import (
	"bytes"

	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/net/html/charset"

	"github.com/oov/mxl2mid/mxl"
)

func main() {
	js.Global.Set("Mxl2mid", conv)
}

func arrayBufferToByteSlice(a *js.Object) []byte {
	return js.Global.Get("Uint8Array").New(a).Interface().([]byte)
}

func conv(in *js.Object, cs string) *js.Object {
	buf := bytes.NewBuffer(arrayBufferToByteSlice(in))
	mxldoc, err := mxl.Decode(buf)
	if err != nil {
		panic(err)
	}

	buf.Reset()
	e, _ := charset.Lookup(cs)
	_, err = mxldoc.MIDI(e.NewEncoder()).WriteTo(buf)
	if err != nil {
		panic(err)
	}
	return js.NewArrayBuffer(buf.Bytes())
}
