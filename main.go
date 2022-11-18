package main

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	processes, err := process.Processes()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to fetch processes: %v\n", err)
		os.Exit(1)
	}

	for _, p := range processes {
		ppid, err := p.Ppid()
		if err != nil {
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
