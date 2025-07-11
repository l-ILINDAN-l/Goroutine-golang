package main

import (
	"fmt"
	"math/big"
)

func main() {
	sA := "23075900000000000000000000"
	sB := "1007202500000000000000000"

	a := new(big.Int)
	b := new(big.Int)

	a.SetString(sA, 10)
	b.SetString(sB, 10)

	fmt.Printf("a: %s\n", a.String())
	fmt.Printf("b: %s\n", b.String())

	sum := new(big.Int)
	sum.Add(a, b)
	fmt.Printf("a + b = %s\n", sum.String())

	diff := new(big.Int)
	diff.Sub(a, b)
	fmt.Printf("a - b = %s\n", diff.String())

	prod := new(big.Int)
	prod.Mul(a, b)
	fmt.Printf("a * b = %s\n", prod.String())

	quot := new(big.Int)
	quot.Quo(a, b)
	fmt.Printf("a / b = %s\n", quot.String())
}
