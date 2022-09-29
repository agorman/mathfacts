package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	var number = flag.Int("number", 0, "math fact to test (number you're multiplying")
	var max = flag.Int("max", 0, "highest multple, usually 10, 11, or 12")
	var duration = flag.Int("duration", 5, "duration of the test in minutes")
	var csvFile = flag.String("csv", "./results.csv", "location to store test results")
	flag.Parse()

	if *number <= 0 {
		panic("number must be greater than 0")
	}

	if *max <= 0 {
		panic("max must be greater than 0")
	}

	if *duration <= 0 {
		panic("duration must be greater than 0")
	}

	start := time.Now()
	rand.Seed(start.UnixNano())

	// create a channel for the result of each answer
	answerc := make(chan bool)

	// start the tests timer based on the duration flag
	t := time.NewTicker(time.Minute * time.Duration(*duration))

	// start the test
	go test(*number, *max, answerc)

	var results []bool

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		select {
		case <-t.C:
			processResults(*csvFile, *number, start, results)
			return
		case <-sig:
			processResults(*csvFile, *number, start, results)
			return
		case b := <-answerc:
			results = append(results, b)
		}
	}
}

func test(number, max int, answerc chan<- bool) {
	// reader will read the users answers from stdin
	reader := bufio.NewReader(os.Stdin)

	// clear the screen at test start
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	for {
		// the next number to multiple
		next := rand.Intn(max + 1)

		// determine which side of the operator the math fact appears
		// and ask the next question
		flip := rand.Intn(100)
		if flip < 50 {
			fmt.Printf("%d x %d: ", number, next)
		} else {
			fmt.Printf("%d x %d: ", next, number)
		}

		// read the answer from stdin
		var answer int
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Print("\nTry again: ")
				continue
			}

			text = strings.Trim(text, "\n")
			answer, err = strconv.Atoi(text)
			if err != nil {
				fmt.Print("\nTry again: ")
				continue
			}

			break
		}

		// clear the screen after each question
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		// print the result and return it on the answer chan
		if answer == number*next {
			fmt.Println("CORRECT\n")
			answerc <- true
		} else {
			fmt.Printf("INCORRECT: %d\n\n", number*next)
			answerc <- false
		}
	}
}

func processResults(csvFile string, number int, start time.Time, results []bool) {
	fmt.Println("\n\nRESULTS")

	var correct int
	for _, res := range results {
		if res {
			correct++
		}
	}
	total := len(results)
	percentCorrect := int(float64(correct) / float64(total) * 100)
	duration := time.Since(start).String()
	correctPerMinute := float64(correct) / time.Since(start).Minutes()
	date := start.Format("2006-01-02")

	fmt.Printf("%d / %d = %d%% in %s: \n", correct, total, percentCorrect, duration)
	fmt.Printf("%f problems per minute\n", correctPerMinute)

	saveResults(csvFile, number, correct, total, percentCorrect, duration, correctPerMinute, date)
}

func saveResults(csvFile string, number, correct, total, percentCorrect int, duration string, correctPerMinute float64, date string) {
	f, err := os.OpenFile(csvFile, os.O_CREATE|os.O_APPEND|os.O_RDONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}

	f, err = os.OpenFile(csvFile, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if len(records) == 0 {
		records = append(records, []string{"date", "fact", "correct", "total", "time", "correct per minute"})
	}

	records = append(records, []string{
		date,
		fmt.Sprintf("%d", number),
		fmt.Sprintf("%d", correct),
		fmt.Sprintf("%d", total),
		duration,
		fmt.Sprintf("%f", correctPerMinute),
	})

	w := csv.NewWriter(f)
	w.WriteAll(records) // calls Flush internally

	if err := w.Error(); err != nil {
		panic(err)
	}
}
