package progressbar

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/mattn/go-runewidth"
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
	pbParam.step = 1
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

func GetMax() int {
	return pbParam.limit
}

func SetStep(step int) {
	pbParam.step = step
}

func Increment() {
	if pbParam.value < pbParam.limit {
		pbParam.value += pbParam.step
	}

	DrawProgressBar()
}

func SetValue(value int) {
	pbParam.value = value
	DrawProgressBar()
}

func Value() int {
	return pbParam.value
}

func Pos() int {
	return pbParam.value * 100 / pbParam.limit
}

func WriteText(text string) {
	if text == "" {
		return
	}
	str := strings.Trim(text, "\t\n ")
	clr := fmt.Sprintf("\r%*s", getWidth()/runewidth.StringWidth(" "), " ")
	fmt.Printf("%s\r%s\n", clr, str)
	DrawProgressBar()
}

func DrawProgressBar() {
	if pbParam.value > pbParam.limit {
		pbParam.value = pbParam.limit
	}
	percent := pbParam.value * 100 / pbParam.limit
	pbwidth := getWidth() - runewidth.StringWidth(fmt.Sprintf("%v/%v [] 100###", pbParam.value, pbParam.limit))
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
