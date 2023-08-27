package main

import (
	"fmt"
	. "github.com/carvermichael/Sim8086/internal"
	"log"
	"os"
)

func main() {

	if len(os.Args) <= 1 {
		log.Panic("you need a fileName, ya dum dum!")
	}

	fileName := os.Args[1]
	fmt.Println(fileName)

	asm_string, _ := GetASMFromFile(fileName)

	fmt.Print(asm_string)

	return
}

