package main

import "fmt"

type Human struct {
	name  string
	age   int
	phone string
}

func (h Human) SayHi() {
	fmt.Printf("Hi, I am %s you can call me on %s\n", h.name, h.phone)
}

type Action struct {
	action string
	Human
}

func main() {
	testAction := Action{

		Human: Human{
			name:  "Test Name",
			age:   19,
			phone: "123456789",
		},
		action: "ring",
	}

	testAction.SayHi()

	fmt.Printf("%s %s\n", testAction.name, testAction.action)
}
