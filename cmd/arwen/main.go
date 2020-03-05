package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ElrondNetwork/arwen-wasm-vm/ipc/arwenpart"
)

func main() {
	fmt.Println("Arwen.main()")

	nodeToArwenFile := os.NewFile(3, "/proc/self/fd/3")
	if nodeToArwenFile == nil {
		log.Fatal("Cannot create file")
	}

	arwenToNodeFile := os.NewFile(4, "/proc/self/fd/4")
	if arwenToNodeFile == nil {
		log.Fatal("Cannot create file")
	}

	// TODO: pass parameters from arguments (blockGaslimit, map of gas, code vmType)
	part, err := arwenpart.NewArwenPart(nodeToArwenFile, arwenToNodeFile)
	if err != nil {
		log.Fatalf("Cannot create ArwenPart: %v", err)
	}

	err = part.StartLoop()
	if err != nil {
		log.Fatalf("Ended Arwen loop: %v", err)
	}
}
