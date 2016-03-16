package main // import "github.com/koenbollen/qm"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const marker = `QM_CHILD`
const timer = 15 * time.Minute

var debug = false

// Serve will open a listensocket on a random port, output it's URL and start
// serving the Open file to every path and method requested.
//
// A timer will be started to close the socket after an constant timeout. Every
// request will reset this timer.
func Serve() {
	content, name, modtime, err := Open(true)
	if err != nil {
		panic(err)
	}

	// Listen on a random port:
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	fmt.Printf("http://%s\n", l.Addr().String())

	t := time.NewTimer(timer)
	go func() {
		<-t.C
		l.Close()
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Reset(timer)
		http.ServeContent(w, r, name, modtime, content)
	})
	fmt.Fprint(os.Stderr, http.Serve(l, nil))
}

// Spawn will start the current binary again with some small changes, it'll
// add a marker in the environ so the child knows it's the child and the open
// files a mocked to provide a simple communication.
func Spawn() {
	env := append(os.Environ(), marker+"=1")

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	defer w.Close()

	files := []*os.File{
		os.Stdin,
		w,
		nil,
	}
	if debug {
		files[2] = os.Stderr
	}

	attr := &os.ProcAttr{
		Env:   env,
		Files: files,
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	}
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		path = os.Args[0]
	}
	process, err := os.StartProcess(path, os.Args, attr)
	if err != nil {
		panic(err)
	}

	// Trap interrupt to close stdin, it's kept open by the child process:
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			os.Stdin.Close()
			process.Kill()
			os.Exit(0)
		}
	}()

	line, _, err := bufio.NewReader(r).ReadLine()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(line))
}

// Open will check the global os.Args and will open the target file or stdin.
// When consume is given as true it'll also consume stdin or a non seekable file
func Open(consume bool) (io.ReadSeeker, string, time.Time, error) {
	var err error

	// Check for using stdin:
	if len(os.Args) < 2 || os.Args[1] == "-" {
		var dat []byte
		if consume {
			dat, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				panic(err)
			}
		}
		return bytes.NewReader(dat), "stdin", time.Now(), nil
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	// File not seekable? Consume and buffer it:
	if _, err := file.Seek(0, os.SEEK_SET); err != nil && consume {
		dat, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		return bytes.NewReader(dat), stat.Name(), stat.ModTime(), nil
	}

	return file, stat.Name(), stat.ModTime(), nil
}

func main() {
	if os.Getenv("DEBUG") != "1" {
		// Prettify panics when not in DEBUG=1 mode:
		defer func() {
			if err := recover(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		}()
	} else {
		debug = true
	}

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("usage: qm [file]")
		return
	}

	Open(false) // try
	if os.Getenv(marker) == "1" {
		Serve()
	} else {
		Spawn()
	}
}
