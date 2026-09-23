package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {

	urlFlag := flag.String("urls", "", "list of started urls")
	depthFlag := flag.Int("depth", 1, "max depth")
	timeoutFlag := flag.Duration("timeout", 2*time.Minute, "overall timeout")
	reqTimeoutFlag := flag.Duration("request-timeout", 10*time.Second, "one request timeout")
	outputFlag := flag.String("output", "result.json", "output result file")
	logFlag := flag.String("log", "logs.log", "log file")

	zagl(*depthFlag, *timeoutFlag, *reqTimeoutFlag, *outputFlag, *logFlag)

	flag.Parse()

	urls := strings.Split(*urlFlag, ",")
	fmt.Println("Started urls:")
	for _, url := range urls {

		fmt.Println(url)

	}

}

func zagl(depthFlag int, timeoutFlag, reqTimeoutFlag time.Duration, outputFlag, logFlag string) {

}
