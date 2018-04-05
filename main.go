package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {
	var duration, interval, rate, size, runs, delay int
	var hold bool
	flag.IntVar(&rate, "rate", 1, "How many log entries per second.")
	flag.IntVar(&size, "size", 128, "How many bytes does one log entry contains (do not use too small sizes).")
	flag.IntVar(&runs, "runs", 1, "How many rounds should it runs.")
	flag.IntVar(&duration, "duration", 5, "How long is one round, in seconds.")
	flag.IntVar(&interval, "interval", 1, "How long should it wait between each round, in seconds.")
	flag.IntVar(&delay, "delay", 0, "How much time to wait before generate logs for the first run, in seconds.")
	flag.BoolVar(&hold, "hold", false, "Determines should it wait the amount of time specified by interval before exits.")
	flag.Parse()

	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}

	pipe := make(chan bool, rate)
	stats := make([][]int, 0, runs)
	run_stats := make([]int, 0, duration)

	run := 0
	count := 0
	start_time := time.Now().UTC()
	log := prepareLogEntry(size, rate, run, duration)

	timeout := time.After(time.Duration(duration) * time.Second)
	one_sec := time.After(time.Second)

mainloop:
	for {
		select {
		case pipe <- true:
			fmt.Printf(log, count)
		case <-one_sec:
			run_stats = append(run_stats, len(pipe))
			pipe = make(chan bool, rate)
			count++
			one_sec = time.After(time.Second)
		case <-timeout:
			if len(run_stats) < cap(run_stats) {
				// this means it didn't get the stats for the last second
				run_stats = append(run_stats, len(pipe))
			}
			stats = append(stats, run_stats)
			run++
			if run >= runs {
				break mainloop
			}
			time.Sleep(time.Duration(interval) * time.Second)

			run_stats = make([]int, 0, duration)
			count = 0
			log = prepareLogEntry(size, rate, run, duration)
			pipe = make(chan bool, rate)
			timeout = time.After(time.Duration(duration) * time.Second)
			one_sec = time.After(time.Second)
		}
	}

	end_time := time.Now().UTC()
	fmt.Printf(
		`{"EventStats":{"startTime":"%v","endTime":"%v","runs":%d,"size":%d,"rate":%d,"duration":%d,"interval":%d,"run_stats":[`, start_time, end_time, runs, size, rate, duration, interval,
	)
	var stats_strings []string
	total := 0
	for _, s := range stats {
		var b strings.Builder
		b.WriteString(`{"eachSecond":[`)
		sum := 0
		last := len(s) - 1
		for i := 0; i < last; i++ {
			sum += s[i]
			b.WriteString(fmt.Sprintf("%d,", s[i]))
		}
		sum += s[last]
		b.WriteString(fmt.Sprintf(`%d], "totalGenerated": %v}`, s[last], sum))
		total += sum
		stats_strings = append(stats_strings, b.String())
	}
	fmt.Printf(`%v],"total": %d}}`, strings.Join(stats_strings, ","), total)

	if hold {
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func prepareLogEntry(size, rate, run, duration int) string {
	var b strings.Builder
	width := len(fmt.Sprintf("%d", duration))
	b.WriteString(fmt.Sprintf("run=%d,count=%%-%dd,size=%d,rate=%d,message=", run, width, size, rate))
	verbSize := len(fmt.Sprintf("%%%dd", width))
	remains := size - b.Len() - width + verbSize
	if remains > 0 {
		c := []byte{'*'}[0]
		for i := 0; i < remains; i++ {
			b.WriteByte(c)
		}
	}
	// if size is small, this will be wrong
	return b.String()[0:size-width+verbSize] + "\n"
}
