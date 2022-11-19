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

func printProcessStats(startTime time.Time) {
	timeOffset := float64(time.Since(startTime)) / float64(time.Second)

	processes, err := process.Processes()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to fetch processes: %v\n", err)
		os.Exit(1)
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

		totalCPU += cpu
		totalMemoryRSS += mem.RSS
		if !*sum {
			name, err := p.Name()
			if err != nil {
				continue
			}

			formatProcess(timeOffset, p.Pid, ppid, mem.RSS, cpu, name)
		}
	}

	formatSummary(timeOffset, totalMemoryRSS, totalCPU)
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
	quit := make(chan bool, 1)

	if *verbose {
		_, _ = fmt.Fprintf(os.Stderr, "Starting process monitoring\n")
	}
	for {
		select {
		case <-ticker.C:
			if checkProcess(mainPid) != nil {
				if *verbose {
					_, _ = fmt.Fprintf(os.Stderr, "Process with pid %d died, exiting\n", mainPid)
				}
				quit <- true
			} else {
				printProcessStats(startTime)
			}
		case <-sigChan:
			if *verbose {
				_, _ = fmt.Fprintf(os.Stderr, "Received termination signal, exiting\n")
			}
			quit <- true
		case <-quit:
			os.Exit(0)
		}
	}
}
