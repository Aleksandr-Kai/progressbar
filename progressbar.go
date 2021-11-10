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

var params struct {
	limit    int
	step     int
	value    int
	interval time.Duration
	current  *time.Timer
	mu       sync.Mutex
	texts    strings.Builder
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func init() {
	params.limit = 0
	params.step = 1
	params.value = 0
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
	if params.interval > time.Second || params.interval < minInterval {
		params.interval = minInterval
	}
	params.current = time.AfterFunc(time.Millisecond, func() {
		params.current.Reset(params.interval)
		DrawProgressBar()
	})
}

func Break() {
	if params.current == nil {
		return
	}
	params.current.Stop()
	fmt.Println()
}

func SetInterval(interval time.Duration) error {
	if interval < minInterval {
		return fmt.Errorf("interval must be greater then %v", minInterval)
	}
	return nil
}

func SetMax(max int) {
	params.limit = max
	if params.current != nil {
		params.current.Stop()
	}
	start()
}

func GetMax() int {
	return params.limit
}

func SetStep(step int) {
	params.step = step
}

func Increment() {
	params.mu.Lock()
	defer params.mu.Unlock()
	if params.value < params.limit {
		params.value += params.step
	}
}

func SetValue(value int) {
	params.mu.Lock()
	defer params.mu.Unlock()
	if value > params.limit {
		return
	}
	params.value = value
}

func Value() int {
	return params.value
}

func Pos() int {
	return params.value * 100 / params.limit
}

func WriteText(text string) {
	if params.current == nil {
		fmt.Println(text)
		return
	}
	if text == "" {
		return
	}

	//sb := strings.Builder{}
	//sb.WriteString(text)
	//w := getWidth()
	//sbw := runewidth.StringWidth(sb.String())
	//for sbw < w {
	//	sb.WriteString(" ")
	//	sbw = runewidth.StringWidth(sb.String())
	//}
	//sb.WriteString("\n")
	params.mu.Lock()
	//params.texts.WriteString(sb.String())
	params.texts.WriteString(text + "\n")
	params.mu.Unlock()
}

func DrawProgressBar() {
	if params.value > params.limit {
		params.value = params.limit
	}
	var percent int
	if params.limit == 0 {
		percent = 100
	} else {
		percent = params.value * 100 / params.limit
	}
	pbwidth := getWidth() - runewidth.StringWidth(fmt.Sprintf("%v/%v [] 100###", params.limit, params.limit))
	l := int(float32(percent) / 100.0 * float32(pbwidth))
	pb := ""
	for ii := 0; ii < l; ii++ {
		pb += "\u2589"
	}
	for pbwidth > runewidth.StringWidth(pb) {
		pb += "\u2591"
	}
	params.mu.Lock()
	//fmt.Printf("\r%s%v/%v [%s] %3d%%  ", params.texts.String(), params.value, params.limit, pb, percent)
	fmt.Printf("\r\u001B[K%s%v/%v [%s] %3d%%  ", params.texts.String(), params.value, params.limit, pb, percent)
	params.texts.Reset()
	params.mu.Unlock()
	if params.value == params.limit {
		Break()
	}
}
