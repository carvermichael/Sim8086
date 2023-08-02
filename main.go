package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	MOV = iota
	ADD
)

const (
	MOV_BITS byte = 0b10001000
	MASK_DW  byte = 0b11111100
	D_BITS   byte = 0b00000010
	W_BITS   byte = 0b00000001
	REG_MASK byte = 0b00111000
	RM_MASK  byte = 0b00000111
)

var w_on_map = map[byte]string{
	0b000: "AX",
	0b001: "CX",
	0b010: "DX",
	0b011: "BX",
	0b100: "SP",
	0b101: "BP",
	0b110: "SI",
	0b111: "DI",
}

var w_off_map = map[byte]string{
	0b000: "AL",
	0b001: "CL",
	0b010: "DL",
	0b011: "BL",
	0b100: "AH",
	0b101: "CH",
	0b110: "DH",
	0b111: "BH",
}

var prog []byte
var err error
var curr int = 0

func main() {

	if len(os.Args) <= 1 {
		log.Panic("you need a fileName, ya dum dum!")
	}

	fileName := os.Args[1]
	fmt.Println(fileName)

	prog, err = ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var builder strings.Builder

	for curr = 0; curr < len(prog); {
		b := prog[curr]

		dw_off := b & MASK_DW

		if (dw_off ^ MOV_BITS) == 0 {
			builder.WriteString("MOV ")

			// get w bit
			w_bit_on := (b & W_BITS) != 0

			curr++
			b2 := prog[curr]

			// REG
			reg_bytes := b2 & REG_MASK
			reg_bytes = reg_bytes >> 3

			var reg_str string
			if w_bit_on {
				reg_str = w_on_map[reg_bytes]
			} else {
				reg_str = w_off_map[reg_bytes]
			}

			// R/M
			rm_bytes := b2 & RM_MASK

			var rm_str string
			if w_bit_on {
				rm_str = w_on_map[rm_bytes]
			} else {
				rm_str = w_off_map[rm_bytes]
			}

			// d_bit == 0 --> REG is NOT the Dest --> R/M REG
			// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
			if (b & D_BITS) == 0 {
				builder.WriteString(fmt.Sprintf("%s %s", rm_str, reg_str))
			} else {
				builder.WriteString(fmt.Sprintf("%s %s", reg_str, rm_str))
			}
			builder.WriteString("\n")
		}

		curr++
	}
	fmt.Print(builder.String())
}
