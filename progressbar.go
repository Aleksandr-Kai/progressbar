package progressbar

import (
	"fmt"
	"github.com/mattn/go-runewidth"
	"strings"
	"syscall"
	"unsafe"
)

var pbParam struct {
	limit int
	step  int
	value int
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func init() {
	pbParam.limit = 0
	pbParam.step = 0
	pbParam.value = 0
}

func getWidth() int {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return int(ws.Col)
}

func SetMax(max int) {
	pbParam.limit = max
}

func SetStep(step int) {
	pbParam.step = step
}

func Increment() {
	pbParam.value += pbParam.step
	Print()
}

func SetValue(value int) {
	pbParam.value = value
	Print()
}

func Value() int {
	return pbParam.value
}

func WriteText(text string) {
	str := strings.Trim(text, "\t\n ")
	for getWidth() > runewidth.StringWidth(str) {
		str += " "
	}
	fmt.Println("\r" + str)
	Print()
}

func Print() {
	if pbParam.value > pbParam.limit {
		return
	}
	percent := pbParam.value * 100 / pbParam.limit
	pbwidth := getWidth() - runewidth.StringWidth(fmt.Sprintf("%v/%v [] 100/%", pbParam.value, pbParam.limit))
	l := int(float32(percent) / 100.0 * float32(pbwidth))
	pb := ""
	for ii := 0; ii < l; ii++ {
		pb += "\u2589"
	}
	for pbwidth > runewidth.StringWidth(pb) {
		pb += "\u2591"
	}
	fmt.Printf("\r%v/%v [%s] %3d%%  ", pbParam.value, pbParam.limit, pb, percent)
}
