package main

import "fmt"

type Reaction interface {
	React()
}

type Cat struct{}

func (c Cat) React() {
	fmt.Println("meow")
}

type Dog struct{}

func (d Dog) Bark() {
	fmt.Println("bark")
}

type DogAdapter struct {
	dog *Dog
}

func (d DogAdapter) React() {
	d.dog.Bark()
}

func AnimalReaction(reaction Reaction) {
	fmt.Println("animal reaction")
	reaction.React()
}

func main() {
	cat := Cat{}
	dog := Dog{}
	AnimalReaction(cat)
	//AnimalReaction(dog) Error
	AnimalReaction(DogAdapter{&dog})
}
