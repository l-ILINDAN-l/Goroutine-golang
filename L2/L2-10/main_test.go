// main_test.go
package main

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"strings"
	"testing"
)

// resetFlags сбрасывает все глобальные переменные флагов к их значениям по умолчанию.
// Это КРИТИЧЕСКИ ВАЖНО для того, чтобы тесты не влияли друг на друга.
func resetFlags() {
	keyColumn = 0
	numSort = false
	reverseSort = false
	uniqSort = false
	monthSort = false
	ignoreBlanks = false
	checkStrings = false
	humanSuffixSort = false
	outputFile = ""
	batchSize = DefaultChunkSize
}

func TestSortLogic(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
		setup    func()
	}{
		{
			name:     "Простая строковая сортировка",
			input:    []string{"c", "a", "b"},
			expected: []string{"a", "b", "c"},
			setup:    func() { /* нет флагов */ },
		},
		{
			name:     "Обратная строковая сортировка (-r)",
			input:    []string{"c", "a", "b"},
			expected: []string{"c", "b", "a"},
			setup:    func() { reverseSort = true },
		},
		{
			name:     "Числовая сортировка (-n)",
			input:    []string{"10", "2", "1"},
			expected: []string{"1", "2", "10"},
			setup:    func() { numSort = true },
		},
		{
			name:     "Числовая сортировка с текстом",
			input:    []string{"10 item", "2 item", "1 item"},
			expected: []string{"1 item", "2 item", "10 item"},
			setup:    func() { numSort = true },
		},
		{
			name:     "Сортировка по 2-й колонке (-k 2)",
			input:    []string{"z 1", "y 3", "x 2"},
			expected: []string{"z 1", "x 2", "y 3"},
			setup:    func() { keyColumn = 2; numSort = true },
		},
		{
			name:     "Сортировка по месяцам (-M)",
			input:    []string{"Mar", "Jan", "Feb"},
			expected: []string{"Jan", "Feb", "Mar"},
			setup:    func() { monthSort = true },
		},
		{
			name:     "Сортировка по месяцам (исправлено)",
			input:    []string{"Mar", "Jan", "Feb"},
			expected: []string{"Jan", "Feb", "Mar"},
			setup:    func() { monthSort = true },
		},
		{
			name:     "Сортировка с учетом суффиксов (-H)",
			input:    []string{"1G", "2K", "3M"},
			expected: []string{"2K", "3M", "1G"},
			setup:    func() { humanSuffixSort = true },
		},
		{
			name:     "Игнорирование хвостовых пробелов (-b)",
			input:    []string{"a  ", "c", "b "},
			expected: []string{"a", "b", "c"},
			setup:    func() { ignoreBlanks = true },
		},
		{
			name:     "Уникальные строки (-u)",
			input:    []string{"c", "a", "c", "b", "a"},
			expected: []string{"a", "b", "c"}, // -u применяется после сортировки
			setup:    func() { uniqSort = true },
		},
		{
			name:     "Уникальные строки по колонке (-u -k 2)",
			input:    []string{"x a", "y b", "z a"},
			expected: []string{"x a", "y b"},
			setup:    func() { uniqSort = true; keyColumn = 2 },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Сбрасываем флаги перед каждым тестом
			resetFlags()
			// Настраиваем флаги для текущего теста
			tc.setup()

			// Копируем входной срез, чтобы не изменять его
			inputCopy := make([]string, len(tc.input))
			copy(inputCopy, tc.input)

			// Логика обработки флага -b (применяется до сортировки)
			if ignoreBlanks {
				for i, line := range inputCopy {
					inputCopy[i] = strings.TrimRight(line, " \t")
				}
			}

			// Создаем функцию сравнения
			lessFunc := newLessStringsFunc()

			// Сортируем
			sort.SliceStable(inputCopy, func(i, j int) bool {
				if reverseSort {
					return lessFunc(inputCopy[j], inputCopy[i])
				}
				return lessFunc(inputCopy[i], inputCopy[j])
			})

			// Логика обработки флага -u (применяется после сортировки)
			if uniqSort {
				inputCopy = getUnique(inputCopy)
			}

			// Проверяем результат
			assert.Equal(t, tc.expected, inputCopy, "результат сортировки не совпал с ожидаемым")
		})
	}
}

// getUnique - вспомогательная функция для -u, которую мы можем протестировать отдельно
func getUnique(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	result := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		val1 := lines[i]
		val2 := lines[i-1]
		if keyColumn > 0 {
			val1 = extractColumn(val1, keyColumn)
			val2 = extractColumn(val2, keyColumn)
		}
		if val1 != val2 {
			result = append(result, lines[i])
		}
	}
	return result
}
