package main

import (
	"fmt"
	"log"

	"github.com/FactomProject/factomd/modules/pubsub2"
	"github.com/FactomProject/factomd/modules/pubsub2/localchannel"
)

func main() {
	local := localchannel.New(5)

	pubsub2.GlobalRegistry().Register("/basic", local)

	go Write()

	Read()
}

func Read() {
	c, ok := pubsub2.GlobalRegistry().Get("/basic")
	if !ok {
		log.Fatalln("channel /basic doesn't exist")
	}

	for v := range c.GetReader().Channel() {
		fmt.Printf("<- %+v\n", v)
	}

	fmt.Println("channel was closed")
}

func Write() {
	c, ok := pubsub2.GlobalRegistry().Get("/basic")
	if !ok {
		log.Fatalln("channel /basic doesn't exist")
	}

	for i := 0; i < 16; i++ {
		c.GetWriter().Write(i)
	}

	c.Close()
}
