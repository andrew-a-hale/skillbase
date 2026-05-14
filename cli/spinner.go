package cli

import (
	"fmt"
	"io"
	"os"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	w    io.Writer
	msg  string
	stop chan struct{}
	done chan struct{}
}

func newSpinner(w io.Writer, msg string) *Spinner {
	s := &Spinner{
		w:    w,
		msg:  msg,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go s.run()
	return s
}

func (s *Spinner) run() {
	defer close(s.done)
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-s.stop:
			clearLine(s.w)
			return
		case <-ticker.C:
			clearLine(s.w)
			fmt.Fprintf(s.w, "%s %s", spinnerFrames[i%len(spinnerFrames)], s.msg)
			i++
		}
	}
}

func (s *Spinner) Stop() {
	select {
	case <-s.stop:
		// already stopped
		return
	default:
		close(s.stop)
		<-s.done
	}
}

func clearLine(w io.Writer) {
	fmt.Fprintf(w, "\r\033[K")
}

func runSpinner(msg string) *Spinner {
	return newSpinner(os.Stderr, msg)
}
