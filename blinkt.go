// MIT License

// Copyright (c) 2017 Alex Ellis
// Copyright (c) 2017 Isaac "Ike" Arias

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE

package blinkt

import (
	"log"
	"math"
	"strconv"
	"time"

	"github.com/ngpitt/gpio"
)

const (
	DAT   = 23
	CLK   = 24
	White = "FFFFFF"
	Red   = "FF0000"
	Green = "00FF00"
	Blue  = "0000FF"
	Off   = "000000"
)

type BlinktObj struct {
	ledSettings []*ledSetting
	gpio        gpio.Gpio
}

type Blinkt interface {
	Set(led int, color string, brightness float64)
	SetAll(color string, brightness float64)
	Flash(led int, color string, brightness float64, times int, duration time.Duration)
	Show()
	Close(color string, brightness float64)
}

type ledSetting struct {
	red   int
	blue  int
	green int
}

func NewBlinkt(color string, brightness float64) Blinkt {
	o := &BlinktObj{
		make([]*ledSetting, 8),
		gpio.NewGpio(),
	}
	for i := 0; i < 8; i++ {
		o.ledSettings[i] = &ledSetting{}
	}
	for i := 3; i >= 0; i-- {
		for j := 0.0; j < brightness; j += brightness / 10 {
			o.Set(i, color, j)
			o.Set(7-i, color, j)
			o.Show()
		}
		o.Set(i, color, brightness)
		o.Set(7-i, color, brightness)
		o.Show()
	}
	o.SetAll(Off, 0)
	o.Show()
	return o
}

func (o *BlinktObj) Set(led int, color string, brightness float64) {
	ls := o.ledSettings[led]
	r, err := strconv.ParseInt(color[:2], 16, 32)
	if err != nil {
		log.Panicln(err.Error())
	}
	g, err := strconv.ParseInt(color[2:4], 16, 32)
	if err != nil {
		log.Panicln(err.Error())
	}
	b, err := strconv.ParseInt(color[4:6], 16, 32)
	if err != nil {
		log.Panicln(err.Error())
	}
	ls.red = int(math.Floor(float64(r)*brightness + 0.5))
	ls.green = int(math.Floor(float64(g)*brightness + 0.5))
	ls.blue = int(math.Floor(float64(b)*brightness + 0.5))
}

func (o *BlinktObj) SetAll(color string, brightness float64) {
	for i := 0; i < 8; i++ {
		o.Set(i, color, brightness)
	}
}

func (o *BlinktObj) Show() {
	o.cycleClock(0, 32)
	for _, ls := range o.ledSettings {
		o.writeInt(255)
		o.writeInt(ls.blue)
		o.writeInt(ls.green)
		o.writeInt(ls.red)
	}
	o.cycleClock(1, 4)
}

func (o *BlinktObj) Flash(led int, color string, brightness float64, times int, duration time.Duration) {
	for i := 0; i < times; i++ {
		o.Set(led, color, brightness)
		o.Show()
		time.Sleep(duration)
		o.Set(led, Off, 0)
		o.Show()
		time.Sleep(duration)
	}
}

func (o *BlinktObj) Close(color string, brightness float64) {
	o.SetAll(color, brightness)
	o.Show()
	for i := 0; i <= 3; i++ {
		for j := brightness; j > 0; j -= brightness / 10 {
			o.Set(i, color, j)
			o.Set(7-i, color, j)
			o.Show()
		}
		o.Set(i, Off, 0)
		o.Set(7-i, Off, 0)
		o.Show()
	}
	o.gpio.Close()
}

func (o *BlinktObj) cycleClock(value int, cycles int) {
	o.gpio.Write(DAT, value)
	for i := 0; i < cycles; i++ {
		o.gpio.Write(CLK, 1)
		o.gpio.Write(CLK, 0)
	}
}

func (o *BlinktObj) writeInt(value int) {
	for i := 0; i < 8; i++ {
		o.gpio.Write(DAT, value&128>>7)
		o.gpio.Write(CLK, 1)
		o.gpio.Write(CLK, 0)
		value <<= 1
	}
}
