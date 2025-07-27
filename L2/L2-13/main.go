package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	fields    string
	delimiter string
	separated bool
)

var rootCmd = &cobra.Command{
	Use:   "cut",
	Short: "cut",
	Long:  "cut",
	Run:   runCut,
}

func init() {
	rootCmd.Flags().StringVarP(&fields, "fields", "f", "", "Specifying the numbers of fields (columns) to be displayed. "+
		"Comma-separated numbers, ranges are possible.")
	rootCmd.Flags().StringVarP(&delimiter, "delimiter", "d", "\t", "Use a different separator (character)")
	rootCmd.Flags().BoolVarP(&separated, "separated", "s", false, "Separate fields")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getInputScanner() *bufio.Scanner {
	return bufio.NewScanner(os.Stdin)
}

func getOutputWriter() (*bufio.Writer, func() error) {
	writer := bufio.NewWriter(os.Stdout)
	return writer, writer.Flush
}

func runCut(_ *cobra.Command, _ []string) {
	if fields == "" {
		_, _ = fmt.Fprintln(os.Stderr, "you must specify a list of bytes, characters, or fields")
		os.Exit(1)
	}
	scanner := getInputScanner()
	writer, flush := getOutputWriter()
	if utf8.RuneCountInString(delimiter) > 1 {
		_, _ = fmt.Fprintln(os.Stderr, "the delimiter must be a single character")
		os.Exit(1)
	}

	split := buildSplit()
	writerFields, err := buildFieldsWriter(writer, flush)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.Contains(text, delimiter) {
			if !separated {
				_, _ = fmt.Fprintln(writer, text)
			}
			continue
		}
		columns := split(text)
		if columns == nil {
			continue
		}
		err = writerFields(columns)
		if err != nil {
			return
		}
	}

}

func buildSplit() func(s string) []string {
	return func(s string) []string {
		if !strings.Contains(s, delimiter) {
			if separated {
				return nil
			}
			return []string{s}
		}
		return strings.Split(s, delimiter)
	}
}

func buildFieldsWriter(writer *bufio.Writer, flush func() error) (func(strs []string) error, error) {
	fieldsFiler, err := fieldsParse()
	if err != nil {
		return nil, err
	}
	return func(strs []string) error {
		filteredStr := make([]string, 0, len(strs)/2)
		for col, str := range strs {
			if fieldsFiler(col) {
				filteredStr = append(filteredStr, str)
			}
		}
		_, err = writer.Write([]byte(strings.Join(filteredStr, delimiter) + "\n"))
		if err != nil {
			return err
		}
		err = flush()
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func fieldsParse() (func(int) bool, error) {

	fs := strings.Split(fields, ",")

	ranges := make([]Range, 0, len(fs)/2)
	nums := make([]int, 0, len(fs)/2)

	for _, f := range fs {
		if strings.Contains(f, "-") {

			borders := strings.Split(f, "-")
			if len(borders) != 2 {
				return nil, errors.New("invalid field range")
			}

			templRange := &Range{}

			if borders[0] == "" && borders[1] == "" {
				return nil, errors.New("invalid range with no endpoint: - ")
			}

			if borders[0] == "" {
				templRange.Left = 1
			} else if leftBorder, err := strconv.Atoi(borders[0]); err == nil {
				templRange.Left = leftBorder
			} else {
				return nil, err
			}

			if borders[1] == "" {
				templRange.Right = math.MaxInt
			} else if rightBorder, err := strconv.Atoi(borders[1]); err == nil {
				templRange.Right = rightBorder
			} else {
				return nil, err
			}

			ranges = append(ranges, *templRange)

		} else {
			num, err := strconv.Atoi(f)
			if err != nil {
				return nil, err
			}
			nums = append(nums, num)
		}
	}

	return func(value int) bool {
		for _, num := range nums {
			if num == value+1 {
				return true
			}
		}
		for _, r := range ranges {
			if r.Left <= value+1 && value+1 <= r.Right {
				return true
			}
		}
		return false
	}, nil
}

// Range is struct for range numbers of fields (columns) to be displayed
type Range struct {
	Left, Right int
}
