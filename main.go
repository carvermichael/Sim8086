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

//const (
//	MOV_BITS byte = 0b10001000
//	MASK_DW  byte = 0b11111100
//	D_BITS   byte = 0b00000010
//	W_BITS   byte = 0b00000001
//	REG_MASK byte = 0b00111000
//	RM_MASK  byte = 0b00000111
//)

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

var prog []byte             // all bytes
var i int                   // current index int = 0
var b byte                  // current byte
var builder strings.Builder // builds ultimate output

var err error

/*
	TODO: Assignment 2:
		- Register to Memory + Memory to Register
		- Immediate to Register
*/

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

	for i = 0; i < len(prog); {
		b = prog[i]

		if (b&0b11111100)^0b10001000 == 0 {
			handleRegToFromRegMem()
		}

		i++
	}
	fmt.Print(builder.String())
}

// MOV #1
func handleRegToFromRegMem() {

	builder.WriteString("MOV ")

	// get d, w, mod, reg, and r/m first
	w_bit_on := (b & 0b00000001) != 0
	d_bit_on := (b & 0b00000010) != 0

	i++
	b2 := prog[i]

	// MOD
	//mod_bits := (b2 & 0b11000000) >> 6

	// REG
	reg_bits := (b2 & 0b00111000) >> 3

	// R/M
	rm_bits := b2 & 0b00000111

	// TODO: the below is just the initial homework case (mod == 11), rework for other mod cases
	var reg_str string
	if w_bit_on {
		reg_str = w_on_map[reg_bits]
	} else {
		reg_str = w_off_map[reg_bits]
	}

	var rm_str string
	if w_bit_on {
		rm_str = w_on_map[rm_bits]
	} else {
		rm_str = w_off_map[rm_bits]
	}

	// d_bit == 0 --> REG is NOT the Dest --> R/M REG
	// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
	if d_bit_on {
		builder.WriteString(fmt.Sprintf("%s %s", reg_str, rm_str))
	} else {
		builder.WriteString(fmt.Sprintf("%s %s", rm_str, reg_str))
	}
	builder.WriteString("\n")
}
