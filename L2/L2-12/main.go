package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	numLinesAfter  int
	numLinesBefore int
	numLinesAround int
	outputCount    bool
	ignoreRegister bool
	invertFilter   bool
	fixedString    bool
	outputNumber   bool
)

var rootCmd = &cobra.Command{
	Use:   "grep",
	Short: "Утилита grep на Go",
	Long:  "Фильтрует строки из файла или стандартного ввода (STDIN), в соответствии шаблону",
	Args:  cobra.RangeArgs(1, 2),
	Run:   runGrep,
}

func init() {
	rootCmd.Flags().IntVarP(&numLinesAfter, "num-lines-after", "A", 0, "Number of lines to print after founded string")
	rootCmd.Flags().IntVarP(&numLinesBefore, "num-lines-before", "B", 0, "Number of lines to print before founded string")
	rootCmd.Flags().IntVarP(&numLinesAround, "num-lines-around", "C", 0, "Number of lines to print around string")
	rootCmd.Flags().BoolVarP(&outputCount, "output-count", "c", false, "Output only count")
	rootCmd.Flags().BoolVarP(&ignoreRegister, "ignore-register", "i", false, "Ignore register")
	rootCmd.Flags().BoolVarP(&invertFilter, "invert-filter", "v", false, "Invert filter")
	rootCmd.Flags().BoolVarP(&fixedString, "fixed-string", "F", false, "Fixed string")
	rootCmd.Flags().BoolVarP(&outputNumber, "output-number", "n", false, "Output number founded string")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runGrep(_ *cobra.Command, args []string) {
	outputWriter, flushWriter, err := getOutputWriter()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		if err := flushWriter(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	inputScanner, closeScanner, err := getInputScanner(args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		if err := closeScanner(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	if outputCount {
		runOutputCount(args[0], outputWriter, inputScanner)
		return
	}

	runFilterString(args[0], outputWriter, flushWriter, inputScanner)

}

func getOutputWriter() (io.Writer, func() error, error) {
	writer := bufio.NewWriter(os.Stdout)
	return writer, writer.Flush, nil
}

func getInputScanner(args []string) (*bufio.Scanner, func() error, error) {
	if len(args) == 1 {
		return bufio.NewScanner(os.Stdin), func() error { return nil }, nil
	}

	file, err := os.Open(args[1])
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewScanner(file), file.Close, nil
}

func runOutputCount(pattern string, output io.Writer, inputScanner *bufio.Scanner) {
	counter := 0
	matcher, err := buildMatcher(pattern)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	for inputScanner.Scan() {
		if matcher(inputScanner.Text()) {
			counter++
		}
	}
	if err := inputScanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	if _, err := output.Write([]byte(strconv.Itoa(counter))); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}

func buildWriter(output io.Writer, flush func() error) func(string, int, bool) error {
	return func(s string, numRow int, matched bool) error {
		if outputNumber {
			if matched {
				_, err := output.Write([]byte(fmt.Sprintf("%d:%s\n", numRow, s)))
				if err != nil {
					return err
				}
			} else {
				_, err := output.Write([]byte(fmt.Sprintf("%d-%s\n", numRow, s)))
				if err != nil {
					return err
				}
			}
		} else {
			_, err := output.Write([]byte(fmt.Sprintf("%s\n", s)))
			if err != nil {
				return err
			}
		}
		if err := flush(); err != nil {
			return err
		}
		return nil
	}
}

func runFilterString(pattern string, output io.Writer, flush func() error, inputScanner *bufio.Scanner) {
	countStringBefore := max(numLinesBefore, numLinesAround)
	countStringAfter := max(numLinesAfter, numLinesAround)

	counterAfter := 0
	queue := NewQueue(countStringBefore)
	counterRow := 1

	matcher, err := buildMatcher(pattern)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	writer := buildWriter(output, flush)

	for inputScanner.Scan() {
		text := inputScanner.Text()
		if matcher(text) {
			lines := queue.Lines()
			if lines != nil {
				for _, line := range lines {
					if err = writer(line.Text, line.Num, false); err != nil {
						_, _ = fmt.Fprintln(os.Stderr, err)
					}
				}
				queue.Clear()
			}

			if err = writer(text, counterRow, true); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
			counterAfter = countStringAfter
		} else {
			if counterAfter > 0 {
				if err = writer(text, counterRow, false); err != nil {
					_, _ = fmt.Fprintln(os.Stderr, err)
				}
				counterAfter--
			} else {
				queue.Add(text, counterRow)
			}
		}
		counterRow++
	}
}

func buildMatcher(pattern string) (func(string) bool, error) {
	if fixedString {
		if ignoreRegister {
			pattern = strings.ToLower(pattern)
			return func(s string) bool {
				return strings.Contains(strings.ToLower(s), pattern) != invertFilter
			}, nil
		}

		return func(s string) bool {
			return strings.Contains(s, pattern) != invertFilter
		}, nil
	}
	if ignoreRegister {
		// Правильный способ: добавить флаг (?i) в шаблон
		pattern = "(?i)" + pattern
	}
	regexPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(s string) bool {
		return regexPattern.MatchString(s) != invertFilter
	}, nil

}

// Line is struct for saving data: text and num row
type Line struct {
	Num  int
	Text string
}

// Queue is struct queue with static size
type Queue struct {
	ringBuffer *ring.Ring
	len        int
	cap        int
}

// NewQueue is function creating object Queue
func NewQueue(cap int) *Queue {
	if cap <= 0 {
		return &Queue{}
	}
	return &Queue{ringBuffer: ring.New(cap),
		len: 0,
		cap: cap,
	}
}

// Add is method to add values (s string) to queue
func (q *Queue) Add(s string, num int) {
	if q.cap == 0 {
		return
	}

	q.ringBuffer.Value = Line{Num: num, Text: s}
	q.ringBuffer = q.ringBuffer.Next()

	if q.len < q.cap {
		q.len++
	}
}

// Lines is method to return value saved in queue. If queue len <= 0 return nil
func (q *Queue) Lines() []Line {
	if q.len <= 0 {
		return nil
	}
	result := make([]Line, q.len)
	tail := q.ringBuffer.Move(-q.len)
	for i := 0; i < q.len; i++ {
		if tail.Value != nil {
			result[i] = tail.Value.(Line)
		}
		tail = tail.Next()
	}
	return result
}

// Clear is method to clear queue
func (q *Queue) Clear() {
	q.len = 0
	for i := 0; i < q.cap; i++ {
		q.ringBuffer.Value = nil
		q.ringBuffer = q.ringBuffer.Next()
	}
}
