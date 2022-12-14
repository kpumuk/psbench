package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

var pidFilter = flag.Int("pid", 0, "filter processes by process pid")
var ppidFilter = flag.Int("ppid", 0, "filter processes by parent process pid")
var selfFilter = flag.Bool("self", false, "monitor psbench process")
var waitDuration = flag.Duration("wait", time.Second, "how many seconds to sleep between iterations")
var sum = flag.Bool("sum", true, "print only summary stats instead of per-process details")
var format = flag.String("format", "text", "output format (one of text, json, csv)")
var verbose = flag.Bool("verbose", false, "print verbose details")

func formatProcess(timeOffset float64, pid, ppid int32, memporyRSS uint64, cpu float64, name string) {
	switch *format {
	case "text":
		fmt.Printf("%d (%d) mem=%d cpu=%.2f name=%q\n", pid, ppid, memporyRSS, cpu, name)
	case "csv":
		fmt.Printf("%.06f,%d,%d,%q,%d,%.2f\n", timeOffset, pid, ppid, name, memporyRSS, cpu)
	case "json":
		fmt.Printf("{\"timestamp\":\"%.06f\",\"type\":\"process\",\"pid\":%d,\"ppid\":%d,\"memory\":%d,\"cpu\":%.2f}\n", timeOffset, pid, ppid, memporyRSS, cpu)
	}
}

func formatSummary(timeOffset float64, totalMemoryRSS uint64, totalCPU float64) {
	switch *format {
	case "text":
		fmt.Printf("Total %.06f: mem=%d cpu=%.2f\n", timeOffset, totalMemoryRSS, totalCPU)
	case "csv":
		fmt.Printf("%.06f,,,,%d,%.2f\n", timeOffset, totalMemoryRSS, totalCPU)
	case "json":
		fmt.Printf("{\"timestamp\":\"%.06f\",\"type\":\"summary\",\"memory\":%d,\"cpu\":%.2f}\n", timeOffset, totalMemoryRSS, totalCPU)
	}
}

func printProcessStats(startTime time.Time) (err error) {
	timeOffset := float64(time.Since(startTime)) / float64(time.Second)

	var processes []*process.Process
	if *pidFilter > 0 {
		proc, err := process.NewProcess(int32(*pidFilter))
		if err != nil {
			return err
		}
		processes = append(processes, proc)
	} else {
		processes, err = process.Processes()
		if err != nil {
			return err
		}
	}

	var totalCPU float64
	var totalMemoryRSS uint64
	for _, p := range processes {
		if *pidFilter > 0 && int(p.Pid) != *pidFilter {
			continue
		}

		ppid, err := p.Ppid()
		if err != nil {
			continue
		}
		if *ppidFilter > 0 && (int(ppid) != *ppidFilter && int(p.Pid) != *ppidFilter) {
			continue
		}

		mem, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		cpu, err := p.CPUPercent()
		if err != nil {
			continue
		}

		if !*sum {
			name, err := p.Name()
			if err != nil {
				continue
			}

			formatProcess(timeOffset, p.Pid, ppid, mem.RSS, cpu, name)
		}

		totalCPU += cpu
		totalMemoryRSS += mem.RSS
	}

	formatSummary(timeOffset, totalMemoryRSS, totalCPU)
	return nil
}

func checkProcess(pid int32) error {
	if pid > 0 {
		_, err := process.NewProcess(pid)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()

	if *selfFilter {
		*pidFilter = os.Getpid()
	}

	var mainPid int32
	if *pidFilter > 0 {
		mainPid = int32(*pidFilter)
	} else if *ppidFilter > 0 {
		mainPid = int32(*ppidFilter)
	}

	startTime := time.Now()
	if *format == "csv" {
		fmt.Printf("timestamp,pid,ppid,name,memory_rss,cpu\n")
	}

	ticker := time.NewTicker(*waitDuration)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	done := make(chan bool)

	printStats := func() {
		if checkProcess(mainPid) != nil {
			if *verbose {
				_, _ = fmt.Fprintf(os.Stderr, "Process with pid %d died, exiting\n", mainPid)
			}
			close(done)
		} else if err := printProcessStats(startTime); err != nil {
			close(done)
		}
	}

	if *verbose {
		_, _ = fmt.Fprintf(os.Stderr, "Starting process monitoring\n")
	}
	printStats()
	for {
		select {
		case <-ticker.C:
			printStats()
		case <-sigChan:
			if *verbose {
				_, _ = fmt.Fprintf(os.Stderr, "Received termination signal, exiting\n")
			}
			close(done)
		case <-done:
			os.Exit(0)
		}
	}
}
