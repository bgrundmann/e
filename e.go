package main

import "github.com/nsf/termbox-go"
import "github.com/bgrundmann/e/buf"
import "github.com/bgrundmann/e/motion"
import "github.com/bgrundmann/e/view"
import "io"
import "os"
import "flag"
import "fmt"
import "log"
import "encoding/json"
import "runtime/pprof"

// AppendFile appends the contents of file to buf.
func AppendFile(buf *buf.Buf, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(buf, f)
	return err
}

type RunMode int
const (
	RunModeRegular RunMode = iota
	RunModeRecord
	RunModeReplay
) 

type commandLineArgs struct {
	runMode RunMode
	recordingFile string // name of the file to record/replay
	cpuprofile string
	initialFiles []string
} 

func parseCommandLine() commandLineArgs {
	var recordFile, replayFile string
	var args commandLineArgs
	flag.StringVar(&recordFile, "record", "", "record all events to file")
	flag.StringVar(&replayFile, "replay", "", "replay all events from file")
	flag.StringVar(&args.cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	args.runMode = RunModeRegular
	if recordFile != "" && replayFile != "" {
		fmt.Fprintf(os.Stderr, "Must specify only one of record/replay!\n")
		flag.PrintDefaults()
		os.Exit(1)
	} else if recordFile != "" {
		args.runMode = RunModeRecord
		args.recordingFile = recordFile
	} else if replayFile != "" {
		args.runMode = RunModeReplay
		args.recordingFile = replayFile
	} 
	args.initialFiles = flag.Args()
	return args
} 

// All init* functions below setup some part of the subsystem and return at least
// a cleanup function that should be run when main exits (via defer).

func initTermbox() func() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)
	return termbox.Close
} 

func initEventSource(args commandLineArgs) (nextEvent func() termbox.Event, cleanup func()) {
	switch args.runMode {
	case RunModeRegular:
		// nothing to be done
		return termbox.PollEvent, func() {}
	case RunModeReplay:
		f, err := os.Open(args.recordingFile)
		if err != nil {
			log.Fatal(err)
		} 
		dec := json.NewDecoder(f)
		return func() termbox.Event {
			var ev *termbox.Event
			if err := dec.Decode(&ev); err != nil {
				log.Fatal(err)
			}
			return *ev
		} , func() {
			f.Close()
		}
	case RunModeRecord:
		f, err := os.OpenFile(args.recordingFile, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0600)
		if err != nil {
			log.Fatal(err)
		}
		enc := json.NewEncoder(f)
		return func() termbox.Event {
			ev := termbox.PollEvent()
			if err := enc.Encode(&ev); err != nil {
				log.Fatal(err)
			} 
			return ev
		}, func() {
			f.Close()
		} 
	default:
		panic("Unknown run mode!")
	} 
} 

func initBufferAndView(v *view.View, args commandLineArgs) func() {
	var b buf.Buf
	b.Init()
	v.Init(&b)
	if len(args.initialFiles) > 0 {
		if err := AppendFile(&b, args.initialFiles[0]); err != nil {
			log.Fatal(err)
		} 
	} 
	return func() {}
} 

func initProfiling(args commandLineArgs) func() {
	if args.cpuprofile != "" {
		f, err := os.Create(args.cpuprofile)
		if err != nil {
			log.Fatal(err)
		} 
		pprof.StartCPUProfile(f)
		return pprof.StopCPUProfile
	} else {
		return func() {}
	} 
} 

func main() {
	args := parseCommandLine()
	cleanup := initTermbox(); defer cleanup()
	nextEvent, cleanup := initEventSource(args); defer cleanup()
	var v view.View
	cleanup = initBufferAndView(&v, args); defer cleanup()
	// not that interested in startup and tear down cost
	// so let's start profiling only now
	cleanup = initProfiling(args); defer cleanup()

mainloop:
	for {
		v.Display()
		switch ev := nextEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyPgdn:
				v.PageDown()
			case termbox.KeyPgup:
				v.PageUp()
			default:
				switch ev.Ch {
				case 'l':
					v.MoveCursor(motion.RuneForward)
				case 'h':
					v.MoveCursor(motion.RuneBackward)
				case 'j':
					v.MoveCursor(motion.LineForward)
				case 'k':
					v.MoveCursor(motion.LineBackward)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
