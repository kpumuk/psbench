# psbench

A very simple process monitoring utility that can measure CPU and memory usage for a given process or a group of processes, and report at requested intervals.

The monitoring will stop as soon as the specified process exits.

> Please note: the tool was implemented to support [On nginx client headers parsing](https://dmytro.sh/blog/on-nginx-client-headers-parsing/) blog post and probably has very little potential on its own. Or maybe it does — let me know.

## Installation

Simplest way to download and compile the tool would be:

```bash
go install github.com/kpumuk/psbench@latest
```

## Usage

```text
Usage of psbench:
  -format string
    	output format (one of text, json, csv) (default "text")
  -pid int
    	filter processes by process pid
  -ppid int
    	filter processes by parent process pid
  -sum
    	print only summary stats instead of per-process details (default true)
  -verbose
    	print verbose details
  -wait duration
    	how many seconds to sleep between iterations (default 1s)
```

## Examples

Print memory and CPU for all processes in a human-readable format:

```text
$ psbench -sum=false
583 (1) mem=82853888 cpu=0.04 name="loginwindow"
909 (1) mem=12025856 cpu=0.09 name="distnoted"
947 (1) mem=63913984 cpu=0.04 name="secd"
...
Total 1.001042: mem=45648916480 cpu=86.60
```

Monitor a single process, print details in CSV format every 0.5 seconds:

```text
$ psbench -wait=500ms -pid=$(cat nginx.pid) -format=csv
timestamp,pid,ppid,name,memory_rss,cpu
0.501697,,,,1376256,0.00
1.001124,,,,1376256,0.00
1.500912,,,,1376256,0.00
```

Monitor a process and all sub-processes, print details in JSON every second:

```text
$ psbench -ppid=$(cat nginx.pid) -format=json -sum=false
{"timestamp":"1.001312","type":"process","pid":45547,"ppid":1,"memory":1376256,"cpu":0.00}
{"timestamp":"1.001312","type":"process","pid":45549,"ppid":45547,"memory":1949696,"cpu":0.00}
{"timestamp":"1.001312","type":"process","pid":45550,"ppid":45547,"memory":1998848,"cpu":0.00}
{"timestamp":"1.001312","type":"process","pid":45551,"ppid":45547,"memory":1966080,"cpu":0.00}
{"timestamp":"1.001312","type":"process","pid":45552,"ppid":45547,"memory":1949696,"cpu":0.00}
{"timestamp":"1.001312","type":"process","pid":45553,"ppid":45547,"memory":1966080,"cpu":0.00}
{"timestamp":"1.001312","type":"process","pid":45554,"ppid":45547,"memory":1933312,"cpu":0.00}
{"timestamp":"1.001312","type":"summary","memory":13139968,"cpu":0.00}
```