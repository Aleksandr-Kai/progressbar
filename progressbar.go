package progressbar

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/mattn/go-runewidth"
)

const (
	minInterval = time.Millisecond * 100
)

var pbParam struct {
	limit    int
	step     int
	value    int
	interval time.Duration
	current  *time.Timer
	mu       sync.Mutex
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

func start() {
	if pbParam.interval > time.Second || pbParam.interval < minInterval {
		pbParam.interval = minInterval
	}
	pbParam.current = time.AfterFunc(time.Millisecond, func() {
		pbParam.current.Reset(pbParam.interval)
		DrawProgressBar()
	})
}

func Break() {
	if pbParam.current == nil {
		return
	}
	pbParam.current.Stop()
	fmt.Println()
}

func SetInterval(interval time.Duration) error {
	if interval < minInterval {
		return fmt.Errorf("interval must be greater then %v", minInterval)
	}
	return nil
}

func SetMax(max int) {
	pbParam.limit = max
	if pbParam.current != nil {
		pbParam.current.Stop()
	}
	start()
}

func GetMax() int {
	return pbParam.limit
}

func SetStep(step int) {
	pbParam.step = step
}

func Increment() {
	pbParam.mu.Lock()
	defer pbParam.mu.Unlock()
	if pbParam.value < pbParam.limit {
		pbParam.value += pbParam.step
	}
}

func SetValue(value int) {
	pbParam.mu.Lock()
	defer pbParam.mu.Unlock()
	if value > pbParam.limit {
		return
	}
	pbParam.value = value
}

func Value() int {
	return pbParam.value
}

func Pos() int {
	return pbParam.value * 100 / pbParam.limit
}

func WriteText(text string) {
	if pbParam.current == nil {
		fmt.Println(text)
		return
	}
	if text == "" {
		return
	}
	str := strings.Trim(text, "\t\n ")
	clr := fmt.Sprintf("\r%*s", getWidth()/runewidth.StringWidth(" "), " ")
	fmt.Printf("%s\r%s\n", clr, str)
	if pbParam.interval > time.Millisecond*500 {
		DrawProgressBar()
	}
}

func DrawProgressBar() {
	if pbParam.value > pbParam.limit {
		pbParam.value = pbParam.limit
	}
	var percent int
	if pbParam.limit == 0 {
		percent = 100
	} else {
		percent = pbParam.value * 100 / pbParam.limit
	}
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
	if pbParam.value == pbParam.limit {
		Break()
	}
}
