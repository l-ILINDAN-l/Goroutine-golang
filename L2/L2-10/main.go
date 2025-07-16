package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	keyColumn       int
	numSort         bool
	reverseSort     bool
	uniqSort        bool
	monthSort       bool
	ignoreBlanks    bool
	checkStrings    bool
	humanSuffixSort bool
	outputFile      string
)

var monthMap = map[string]int{
	"jan": 1,
	"feb": 2,
	"mar": 3,
	"apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"oct": 10,
	"nov": 11,
	"dec": 12,
}

var humanSuffixMap = map[rune]*big.Int{
	'K': big.NewInt(1 << 10),
	'M': big.NewInt(1 << 20),
	'G': big.NewInt(1 << 30),
	'T': big.NewInt(1 << 40),
	'P': big.NewInt(1 << 50),
	'E': new(big.Int).Lsh(big.NewInt(1), 60),
	'Z': new(big.Int).Lsh(big.NewInt(1), 70),
	'Y': new(big.Int).Lsh(big.NewInt(1), 80),
}

var rootCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sort golang application",
	Long:  "Sort golang application rows from file or standard input with big data support",
	Args:  cobra.MaximumNArgs(1),
	Run:   runSort,
}

func init() {
	rootCmd.Flags().IntVarP(&keyColumn, "key", "k", 0, "sort by N-th column")
	rootCmd.Flags().BoolVarP(&numSort, "numeric", "n", false, "sort by numeric value")
	rootCmd.Flags().BoolVarP(&reverseSort, "reverse", "r", false, "sort in reverse order")
	rootCmd.Flags().BoolVarP(&uniqSort, "unique", "u", false, "do not output duplicate lines")

	rootCmd.Flags().BoolVarP(&monthSort, "month", "M", false, "sort by month name")
	rootCmd.Flags().BoolVarP(&ignoreBlanks, "ignore-blanks", "b", false, "ignore tail spaces")
	rootCmd.Flags().BoolVarP(&checkStrings, "check", "c", false, "check if the data is sorted")
	rootCmd.Flags().BoolVarP(&humanSuffixSort, "human-suffix", "h", false, "sort by numeric value, taking into account suffixes (K, M, G, T)")

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "write the result to a file - standard output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runSort(_ *cobra.Command, args []string) {
	outputWriter, closeWriter, err := getOutputWriter()
	defer func() {
		if err := closeWriter(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	lines, err := readLines(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading lines: %v", err)
	}

	if ignoreBlanks {
		trimBlanks(&lines)
	}

	primalLessFunc := newLessFunc(lines)

	var lessFunc func(i, j int) bool
	if reverseSort {
		lessFunc = func(i, j int) bool {
			return primalLessFunc(j, i)
		}
	} else {
		lessFunc = primalLessFunc
	}

	if checkStrings {
		for i := 1; i < len(lines); i++ {
			if lessFunc(i, i-1) {
				_, err = outputWriter.Write([]byte(fmt.Sprintf("sort: -:%d: disorder:%s)", i, lines[i])))
				if err != nil {
					_, _ = fmt.Fprintln(os.Stderr, err)
				}
			}
		}
		return
	} else {
		sort.SliceStable(lines, lessFunc)
	}

	if uniqSort {
		if len(lines) > 0 {
			uniqLines := make([]string, 0)
			lastLine := lines[0]
			uniqLines = append(uniqLines, lastLine)
			for i := 1; i < len(lines); i++ {
				if lastLine != lines[i] {
					uniqLines = append(uniqLines, lines[i])
					lastLine = lines[i]
				}
			}
			lines = uniqLines
		}
	}

	for _, line := range lines {
		_, err = outputWriter.Write([]byte(line + "\n"))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}
}

func getOutputWriter() (io.Writer, func() error, error) {
	if outputFile == "" {
		return bufio.NewWriter(os.Stdout), func() error { return nil }, nil
	}

	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return nil, nil, err
	}
	outputWriter := bufio.NewWriter(file)
	closeWriter := func() error {
		return file.Close()
	}
	return outputWriter, closeWriter, nil
}

func readLines(args []string) ([]string, error) {
	reader, closeReader, err := getInputReader(args)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = closeReader()
	}()

	var lines []string

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil

}

func getInputReader(args []string) (io.Reader, func() error, error) {
	if len(args) == 0 {
		return bufio.NewReader(os.Stdin), func() error { return nil }, nil
	}

	file, err := os.Open(args[0])
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewReader(file), func() error { return file.Close() }, nil
}

func newLessFunc(lines []string) func(i int, j int) bool {
	return func(i, j int) bool {
		line1 := lines[i]
		line2 := lines[j]

		if keyColumn > 0 {
			line1 = extractColumn(line1, keyColumn)
			line2 = extractColumn(line2, keyColumn)
		}

		switch {
		case monthSort:
			numMonth1 := extractMonth(line1)
			numMonth2 := extractMonth(line2)
			return numMonth1 < numMonth2
		case numSort:
			num1 := extractNumeric(line1)
			num2 := extractNumeric(line2)
			return num1 < num2
		case humanSuffixSort:
			size1, _ := parseHumanNumeric(line1)
			size2, _ := parseHumanNumeric(line2)
			return size1.Cmp(size2) == -1
		}
		return line1 < line2
	}
}

func trimBlanks(lines *[]string) {
	for i, line := range *lines {
		(*lines)[i] = strings.TrimRight(line, " \t")
	}
}

func extractColumn(line string, keyColumn int) string {
	fields := strings.Fields(line)

	if keyColumn > 0 && keyColumn <= len(fields) {
		return fields[keyColumn-1]
	} else {
		return ""
	}
}

func extractMonth(line string) int {
	if len(line) < 3 {
		return 0
	}
	month := line[:3]
	return monthMap[month]
}

func extractNumeric(line string) float64 {
	numeric, _ := strconv.ParseFloat(line, 64)
	return numeric
}

func parseHumanNumeric(line string) (*big.Int, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return big.NewInt(0), nil
	}

	var multiplier = big.NewInt(1)
	numPart := line

	lastChar := rune(line[len(line)-1])
	if m, ok := humanSuffixMap[lastChar]; ok {
		multiplier = m
		numPart = line[:len(line)-1]
	}

	f, _, err := big.ParseFloat(numPart, 10, 0, big.ToNearestEven)
	if err != nil {
		return big.NewInt(0), err
	}

	f.Mul(f, new(big.Float).SetInt(multiplier))

	result, _ := f.Int(nil)
	return result, nil
}
