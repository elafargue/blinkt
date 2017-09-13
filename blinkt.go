// MIT License

// Copyright (c) 2017 Alex Ellis
// Copyright (c) 2017 Isaac "Ike" Arias
// Copyright (c) 2017 Nicholas Pitt

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

var gamma = []int{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2,
	2, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5,
	5, 6, 6, 6, 6, 7, 7, 7, 7, 8, 8, 8, 9, 9, 9, 10,
	10, 10, 11, 11, 11, 12, 12, 13, 13, 13, 14, 14, 15, 15, 16, 16,
	17, 17, 18, 18, 19, 19, 20, 20, 21, 21, 22, 22, 23, 24, 24, 25,
	25, 26, 27, 27, 28, 29, 29, 30, 31, 32, 32, 33, 34, 35, 35, 36,
	37, 38, 39, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 50,
	51, 52, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 66, 67, 68,
	69, 70, 72, 73, 74, 75, 77, 78, 79, 81, 82, 83, 85, 86, 87, 89,
	90, 92, 93, 95, 96, 98, 99, 101, 102, 104, 105, 107, 109, 110, 112, 114,
	115, 117, 119, 120, 122, 124, 126, 127, 129, 131, 133, 135, 137, 138, 140, 142,
	144, 146, 148, 150, 152, 154, 156, 158, 160, 162, 164, 167, 169, 171, 173, 175,
	177, 180, 182, 184, 186, 189, 191, 193, 196, 198, 200, 203, 205, 208, 210, 213,
	215, 218, 220, 223, 225, 228, 231, 233, 236, 239, 241, 244, 247, 249, 252, 255}

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
	ls.red = hexToColor(color[:2], brightness)
	ls.green = hexToColor(color[2:4], brightness)
	ls.blue = hexToColor(color[4:6], brightness)
}

func (o *BlinktObj) SetAll(color string, brightness float64) {
	for i := 0; i < 8; i++ {
		o.Set(i, color, brightness)
	}
}

func (o *BlinktObj) Show() {
	o.write(0, 32)
	for _, ls := range o.ledSettings {
		o.writeInt(255)
		o.writeInt(ls.blue)
		o.writeInt(ls.green)
		o.writeInt(ls.red)
	}
	o.write(1, 32)
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

func (o *BlinktObj) write(value, times int) {
	o.gpio.Write(DAT, value)
	for i := 0; i < times; i++ {
		o.gpio.Write(CLK, 1)
		o.gpio.Write(CLK, 0)
	}
}

func (o *BlinktObj) writeInt(value int) {
	for i := 0; i < 8; i++ {
		o.write(value<<uint(i)&128>>7, 1)
	}
}

func hexToColor(hex string, brightness float64) int {
	i, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		log.Panicln(err.Error())
	}
	return gamma[int(math.Floor(float64(i)*brightness+0.5))]
}
