package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

var pidFilter = flag.Int("pid", 0, "filter processes by process pid")
var ppidFilter = flag.Int("ppid", 0, "filter processes by parent process pid")

func main() {
	flag.Parse()

	processes, err := process.Processes()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to fetch processes: %v\n", err)
		os.Exit(1)
	}

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

		name, err := p.Name()
		if err != nil {
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
		fmt.Printf("%d\t%d\t%s\t%d\t%.2f\n", p.Pid, ppid, name, mem.RSS, cpu)
	}
}
