package main

import (
	"bufio"
	"container/heap"
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
	batchSize       int
)

// DefaultChunkSize is const value of chunk size
const DefaultChunkSize = 100000

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
	rootCmd.Flags().BoolVarP(&humanSuffixSort, "human-suffix", "H", false, "sort by numeric value, taking into account suffixes (K, M, G, T)")

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "write the result to a file - standard output")
	rootCmd.Flags().IntVar(&batchSize, "batch-size", DefaultChunkSize, "batch size")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runSort(_ *cobra.Command, args []string) {
	outputWriter, closeWriter, err := getOutputWriter()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		if err = closeWriter(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	inputScanner, closeScanner, err := getInputScanner(args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		if err = closeScanner(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()

	if checkStrings {
		runCheckString(inputScanner, outputWriter)
		return
	}

	runSortString(inputScanner, outputWriter)
}

func runCheckString(inputScanner *bufio.Scanner, outputWriter io.Writer) {
	lessFunc := newLessStringsFuncReverseFlag()

	var prevLine string
	lineNumber := 1

	for inputScanner.Scan() {
		currentLine := inputScanner.Text()
		if ignoreBlanks {
			currentLine = strings.TrimRight(currentLine, " \t")
		}
		if lineNumber != 1 {
			if lessFunc(currentLine, prevLine) {
				if _, err := outputWriter.Write([]byte(fmt.Sprintf("sort: -:%d: disorder: %s", lineNumber, currentLine))); err != nil {
					_, _ = fmt.Fprintln(os.Stderr, err)
				}
				return
			}
		}
		prevLine = currentLine
		lineNumber++
	}
	if err := inputScanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runSortString(inputScanner *bufio.Scanner, outputWriter io.Writer) {
	lessFunc := newLessStringsFuncReverseFlag()
	chunks, err := sortAndSaveChunks(inputScanner, lessFunc)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = writeSortedChunks(outputWriter, chunks, lessFunc); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func getOutputWriter() (io.Writer, func() error, error) {
	if outputFile == "" {
		writer := bufio.NewWriter(os.Stdout)
		return writer, writer.Flush, nil
	}

	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return nil, nil, err
	}
	outputWriter := bufio.NewWriter(file)
	closeWriter := func() error {
		if err = outputWriter.Flush(); err != nil {
			_ = file.Close()
			return err
		}
		return file.Close()
	}
	return outputWriter, closeWriter, nil
}

func getInputScanner(args []string) (*bufio.Scanner, func() error, error) {
	if len(args) == 0 {
		return bufio.NewScanner(os.Stdin), func() error { return nil }, nil
	}
	file, err := os.Open(args[0])
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewScanner(file), file.Close, nil
}

func sortAndSaveChunks(inputScanner *bufio.Scanner, lessFunc func(s1 string, s2 string) bool) ([]string, error) {
	chunkStrings := make([]string, 0)
	var chunks []string
	for inputScanner.Scan() {
		chunkString := inputScanner.Text()
		if ignoreBlanks {
			chunkString = strings.TrimRight(chunkString, " \t")
		}
		chunkStrings = append(chunkStrings, chunkString)
		if len(chunkStrings) >= batchSize {
			fileName, err := saveChunk(chunkStrings, lessFunc)
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, fileName)
			chunkStrings = make([]string, 0)
		}
	}
	if len(chunkStrings) > 0 {
		fileName, err := saveChunk(chunkStrings, lessFunc)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, fileName)
	}
	return chunks, nil
}

func saveChunk(chunkStrings []string, lessFunc func(s1 string, s2 string) bool) (string, error) {
	file, err := os.CreateTemp("", "sort-chunk-*.txt")
	if err != nil {
		return "", err
	}
	sort.SliceStable(chunkStrings, func(i, j int) bool {
		return lessFunc(chunkStrings[i], chunkStrings[j])
	})
	for _, str := range chunkStrings {
		if _, err = file.WriteString(fmt.Sprintf("%s\n", str)); err != nil {
			return "", err
		}
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func writeSortedChunks(outputWriter io.Writer, chunks []string, lessFunc func(s1 string, s2 string) bool) error {
	files := make([]*os.File, len(chunks))
	scanners := make([]*bufio.Scanner, len(chunks))

	defer func() {
		for _, file := range files {
			_ = file.Close()
		}
	}()
	for i, chunk := range chunks {
		file, err := os.Open(chunk)
		if err != nil {
			return err
		}
		files[i] = file
		scanners[i] = bufio.NewScanner(file)
	}
	pq := &PriorityQueue{
		lessFunc: lessFunc,
		items:    make([]*Item, 0),
	}

	heap.Init(pq)

	for _, scanner := range scanners {
		scanner.Scan()
		heap.Push(pq, &Item{
			currentString: scanner.Text(),
			scanner:       scanner,
		})
	}

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item)
		_, err := outputWriter.Write([]byte(fmt.Sprintf("%s\n", item.currentString)))
		if err != nil {
			return err
		}
		if item.scanner.Scan() {
			heap.Push(pq, &Item{
				currentString: item.scanner.Text(),
				scanner:       item.scanner,
			})
		}

	}
	return nil
}

func newLessStringsFuncReverseFlag() func(s1, s2 string) bool {
	primalLessFunc := newLessStringsFunc()
	if reverseSort {
		return func(s1, s2 string) bool {
			return primalLessFunc(s2, s1)
		}
	}
	return primalLessFunc
}

func newLessStringsFunc() func(s1, s2 string) bool {
	return func(s1, s2 string) bool {
		if keyColumn > 0 {
			s1 = extractColumn(s1, keyColumn)
			s2 = extractColumn(s2, keyColumn)
		}

		switch {
		case monthSort:
			numMonth1 := extractMonth(s1)
			numMonth2 := extractMonth(s2)
			return numMonth1 < numMonth2
		case numSort:
			var num1, num2 float64
			_, _ = fmt.Sscanf(strings.TrimSpace(s1), "%f", &num1)
			_, _ = fmt.Sscanf(strings.TrimSpace(s2), "%f", &num2)
			return num1 < num2
		case humanSuffixSort:
			size1, _ := parseHumanNumeric(s1)
			size2, _ := parseHumanNumeric(s2)
			return size1.Cmp(size2) == -1
		}
		return s1 < s2
	}
}

func extractColumn(line string, keyColumn int) string {
	fields := strings.Fields(line)

	if keyColumn > 0 && keyColumn <= len(fields) {
		return fields[keyColumn-1]
	}

	return ""
}

func extractMonth(line string) int {
	if len(line) < 3 {
		return 0
	}
	month := strings.ToLower(line[:3])
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

// Item is structure to contain currentString and scanner
type Item struct {
	currentString string
	scanner       *bufio.Scanner
}

// PriorityQueue is structure to heap sort items
type PriorityQueue struct {
	items    []*Item
	lessFunc func(s1, s2 string) bool
}

func (pq PriorityQueue) Len() int {
	return len(pq.items)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq.lessFunc(pq.items[i].currentString, pq.items[j].currentString)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push is method PriorityQueue for push element to heap
func (pq *PriorityQueue) Push(x any) {
	item := x.(*Item)
	pq.items = append(pq.items, item)
}

// Pop is method PriorityQueue for pop element from heap
func (pq *PriorityQueue) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[0 : n-1]
	return item
}
