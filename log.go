package ecs

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
)

type ProgressTracker struct {
	state int
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{state: 0}
}

func (p *ProgressTracker) Info(format string, a ...interface{}) {
	b := []interface{}{green("INFO")}
	b = append(b, a...)

	if p.state > 0 {
		fmt.Printf("\033[1A\033[100D")
	}

	var v string
	switch p.state % 4 {
	case 0:
		v = "|"
	case 1:
		v = "/"
	case 2:
		v = "-"
	case 3:
		v = "\\"
	}

	fmt.Printf("[%s ] "+format+" "+v+"\n", b...)
	p.state++
}

func Text(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

func Info(format string, a ...interface{}) {
	b := []interface{}{green("INFO")}
	b = append(b, a...)
	fmt.Printf("[%s ] "+format+"\n", b...)
}

func Warn(format string, a ...interface{}) {
	b := []interface{}{yellow("WARN")}
	b = append(b, a...)
	fmt.Printf("[%s ] "+format+"\n", b...)
}

func Error(format string, a ...interface{}) {
	b := []interface{}{yellow("ERROR")}
	b = append(b, a...)
	fmt.Printf("[%s] "+format+"\n", b...)
}
