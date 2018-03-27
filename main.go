package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {
	var duration, interval, rate, size, runs int
	flag.IntVar(&rate, "rate", 1, "How many log entries per second.")
	flag.IntVar(&size, "size", 128, "How many bytes does one log entry contains.")
	flag.IntVar(&runs, "runs", 1, "How many rounds should it runs.")
	flag.IntVar(&duration, "duration", 5, "How long is one round, in seconds.")
	flag.IntVar(&interval, "interval", 1, "How long should it wait between each round, in seconds.")
	flag.Parse()

	pipe := make(chan bool, rate)
	stats := make([][]int, 0, runs)
	run_stats := make([]int, 0, duration)

	run := 0
	start_time := time.Now().UTC()
	log := prepareLogEntry(start_time, size, rate, run)

	timeout := time.After(time.Duration(duration) * time.Second)
	one_sec := time.After(time.Second)

mainloop:
	for {
		select {
		case pipe <- true:
			fmt.Println(log)
		case <-one_sec:
			run_stats = append(run_stats, len(pipe))
			pipe = make(chan bool, rate)
			one_sec = time.After(time.Second)
		case <-timeout:
			count := len(pipe)
			if count > 0 {
				run_stats = append(run_stats, count)
			}
			stats = append(stats, run_stats)
			run++
			if run >= runs {
				break mainloop
			}
			run_stats = make([]int, 0, duration)
			time.Sleep(time.Duration(interval) * time.Second)

			log = prepareLogEntry(start_time, size, rate, run)
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
}

func prepareLogEntry(start_time time.Time, size, rate, run int) string {
	var b strings.Builder
	b.WriteString(start_time.Format(time.RFC1123Z))
	b.WriteString(fmt.Sprintf(" run=%d,size=%d,rate=%d,message=", run, size, rate))
	remains := size - b.Len()
	if remains > 0 {
		c := []byte{'*'}[0]
		for i := 0; i < remains; i++ {
			b.WriteByte(c)
		}
		return b.String()
	} else {
		return b.String()[0:size]
	}
}
