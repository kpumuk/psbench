package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

var pidFilter = flag.Int("pid", 0, "filter processes by process pid")
var ppidFilter = flag.Int("ppid", 0, "filter processes by parent process pid")
var waitDuration = flag.Duration("wait", time.Second, "how many seconds to sleep between iterations")
var sum = flag.Bool("sum", true, "print only summary stats instead of per-process details")

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

			fmt.Printf("%d (%d) mem=%d cpu=%.2f name=%q\n", p.Pid, ppid, mem.RSS, cpu, name)
		}
	}

	fmt.Printf("Total %.06f: mem=%d cpu=%.2f\n", timeOffset, totalMemoryRSS, totalCPU)
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
	for {
		if mainPid > 0 {
			_, err := process.NewProcess(mainPid)
			if err != nil {
				break
			}
		}

		printProcessStats(startTime)
		time.Sleep(*waitDuration)
	}
}
