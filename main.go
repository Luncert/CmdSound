package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"flag"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/bobertlo/go-mpg123/mpg123"
	"github.com/gordonklaus/portaudio"
)

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func run(cmd string) string {
	tmp := strings.Split(cmd, " ")
	if len(tmp) == 0 {
		panic(fmt.Errorf("empty command"))
	}
	c := exec.Command(tmp[0], tmp[1:]...)
	out, err := c.Output()
	chk(err)
	return string(out)
}

func print(out []int16) {
	min, max := out[0], out[0]
	for _, i := range out {
		if i > max {
			max = i
		}
		if i < min {
			min = i
		}
	}
	if max-min <= 0 {
		return
	}

	C.system(C.CString("clear"))

	rate := float64(height) / float64(max-min)
	space := float64(len(out)) / float64(width)
	values := make([]int16, width)
	for i := int16(0); i < width; i++ {
		values[i] = int16((float64(out[int16(float64(i)*space)] - min)) * rate)
	}
	for i := int16(0); i < height; i++ {
		for j := int16(0); j < width; j++ {
			if height-values[j] <= i {
				fmt.Print("_")
			}
		}
		fmt.Print("\n")
	}
}

var (
	width  int16 = 100
	height int16 = 20
)

func main() {
	musicPath := flag.String("music", "", "Input the music's path")
	flag.Parse()
	if *musicPath == "" {
		fmt.Println("Usage: -name=$music_path")
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// create mpg123 decoder instance
	decoder, err := mpg123.NewDecoder("")
	chk(err)

	chk(decoder.Open(*musicPath))
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()

	// make sure output format does not change
	decoder.FormatNone()
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

	portaudio.Initialize()
	defer portaudio.Terminate()

	out := make([]int16, 8192)
	stream, err := portaudio.OpenDefaultStream(0, channels, float64(rate), len(out), &out)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	defer stream.Stop()

	go func() {
		for {
			print(out)
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
	}()

	for {
		audio := make([]byte, 2*len(out))
		_, err = decoder.Read(audio)
		if err == mpg123.EOF {
			break
		}
		chk(err)
		chk(binary.Read(bytes.NewBuffer(audio), binary.LittleEndian, out))
		chk(stream.Write())

		select {
		case <-sig:
			return
		default:
		}
	}
}
