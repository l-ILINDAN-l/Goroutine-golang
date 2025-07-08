package main

import (
	"fmt"
	"log"
)

func setByte(number int64, bitNumber int, valueBit byte) (int64, error) {
	if valueBit == 1 {
		return number | (1 << (bitNumber - 1)), nil
	} else if valueBit == 0 {
		return number &^ (1 << (bitNumber - 1)), nil
	} else {
		return number, fmt.Errorf("value bit must be 0 or 1: %d", valueBit)
	}
}

func main() {
	var inputNum int64
	if _, err := fmt.Scan(&inputNum); err != nil {
		log.Fatalf("failed to read input: %v", err)
		return
	}
	fmt.Printf("Your number is %b\n", inputNum)
	res1, err := setByte(inputNum, 2, 0)
	if err != nil {
		log.Fatalf("failed to set byte: %v", err)
		return
	}
	fmt.Printf("Your number with set second bit 0 is %b\n", res1)
	res2, err := setByte(inputNum, 3, 1)
	if err != nil {
		log.Fatalf("failed to set byte: %v", err)
		return
	}
	fmt.Printf("Your number with set third bit 1 is %b\n", res2)

}
