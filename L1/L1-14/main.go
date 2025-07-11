package main

import (
	"fmt"
	"reflect"
)

func typeDefinitionByInterface(iface interface{}) string {
	switch iface.(type) {
	case string:
		return "string"
	case int:
		return "int"
	case bool:
		return "bool"
	case chan int:
		return "chan int"
	case chan string:
		return "chan string"
	case chan bool:
		return "chan bool"
	default:
		return "default"
	}
}

func typeDefinitionByReflect(iface interface{}) string {
	return reflect.TypeOf(iface).String()
}

func main() {
	ci := make(chan int)
	fmt.Println(typeDefinitionByReflect(ci))
	fmt.Println(typeDefinitionByInterface(ci))

	vi := 1
	fmt.Println(typeDefinitionByReflect(vi))
	fmt.Println(typeDefinitionByInterface(vi))

	vs := "var"
	fmt.Println(typeDefinitionByReflect(vs))
	fmt.Println(typeDefinitionByInterface(vs))

	vb := true
	fmt.Println(typeDefinitionByReflect(vb))
	fmt.Println(typeDefinitionByInterface(vb))
}
