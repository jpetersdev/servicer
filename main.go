package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var name = flag.String("cmd", "./server.linux", "the process you want to service")

func main() {
	flag.Parse()
	logger, f := newLogger()
	defer f.Close()
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		_, min, sec := time.Now().Clock()
		remaining := 3600 - ((min * 60) + sec)
		timer := time.NewTimer(time.Second  * time.Duration(remaining))
		logger.Printf("Starting Process %v\n", *name)
		process, stdout, err := start(*name)
		if err != nil {
			logger.Fatal(err)
		}
		go printOutput(stdout, logger)
		select {
		case <-timer.C:
			logger.Printf("Interupting Process %v\n", *name)
			if err := interrupt(process); err != nil {
				logger.Fatal(err)
			}
		case <-s:
			logger.Println("Servicer Shutdown Initiated")
			logger.Printf("Interupting Process %v\n", *name)
			if err := interrupt(process); err != nil {
				logger.Fatal(err)
			}
			return
		}
	}
}

func start(name string) (process *exec.Cmd, stdout io.ReadCloser, err error) {
	process = exec.Command(name)

	stdout, err = process.StdoutPipe()
	if err != nil {return nil, nil, err}

	err = process.Start()
	if err != nil {return nil, nil, err}

	return
}

func interrupt(process *exec.Cmd) error {
	if err := process.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	if _, err := process.Process.Wait(); err != nil {
		return err
	}
	return nil
}

func printOutput(reader io.ReadCloser, loggers ...*log.Logger) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		if len(loggers) != 0 {
			loggers[0].Println(m)
		} else {
			log.Println(m)
		}
	}
}

func newLogger() (*log.Logger, *os.File) {
	f, err := os.OpenFile("errors.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)
	return log.New(wrt, "", log.Lmsgprefix), f
}
