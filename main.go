package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// TODO: Where you left off --> REGMEMTOFROMEITHER and ImmediateWRegMem seem
//		fine with ADD,SUB,CMP --> do Immedate with accumulator (do MOV as well from
//		the last challenge problem???
// 		--> Then do the jumps.

// TODO: this could be a much faster lookup as an array with the index
//			being ((w << 3) | rem)
// 			Perf test to compare?
var w_on_map = map[byte]string{
	0b000: "ax",
	0b001: "cx",
	0b010: "dx",
	0b011: "bx",
	0b100: "sp",
	0b101: "bp",
	0b110: "si",
	0b111: "di",
}

var w_off_map = map[byte]string{
	0b000: "al",
	0b001: "cl",
	0b010: "dl",
	0b011: "bl",
	0b100: "ah",
	0b101: "ch",
	0b110: "dh",
	0b111: "bh",
}

// TODO: index into an array here
var rm_effective_map = map[byte]string{
	0b000: "bx + si",
	0b001: "bx + di",
	0b010: "bp + si",
	0b011: "bp + di",
	0b100: "si",
	0b101: "di",
	0b110: "bp",
	0b111: "bx",
}

var prog []byte             // all bytes
var i int                   // current index int = 0
var b, b2, b3, b4 byte      // current bytes (TODO: array?)
var builder strings.Builder // builds ultimate output

var err error

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
			movRegToFromRegMem()
		} else if (b&0b11110000)^0b10110000 == 0 {
			movImmediateToReg()
		} else if (b&0b11111100)^0b10000000 == 0 {
			// immediate to register/mem == 0b100000XX in first byte
			arithmeticImmediateToReg()
		} else if (b&0b11000100)^0b00000000 == 0 {
			// --> reg/mem and register == 0b00XXX0XX
			arithmeticRegToFromRegMem()
		} else if (b&0b11000100)^0b00000100 == 0 {
			// --> immediate from accumulator == 0b00XXX1XX
			arithmeticImmediateFromAccum()
		}

		i++
		builder.WriteString("\n")
	}
	fmt.Print(builder.String())
}

// MOV #1
func movRegToFromRegMem() {
	regMemToFromEither("mov")
}

func regMemToFromEither(opCode string) {
	builder.WriteString(fmt.Sprintf("%s ", opCode))

	// get d, w, mod, reg, and r/m first
	w_bit_on := (b & 0b00000001) != 0
	d_bit_on := (b & 0b00000010) != 0

	i++
	b2 = prog[i]

	// MOD
	mod_bits := (b2 & 0b11000000) >> 6

	// REG
	reg_bits := (b2 & 0b00111000) >> 3

	// R/M
	rm_bits := b2 & 0b00000111

	var reg_str string
	if w_bit_on {
		reg_str = w_on_map[reg_bits]
	} else {
		reg_str = w_off_map[reg_bits]
	}

	var rm_str string

	// Register Mode (Register to Register)
	if mod_bits == 0b11 {
		if w_bit_on {
			rm_str = w_on_map[rm_bits]
		} else {
			rm_str = w_off_map[rm_bits]
		}

		// d_bit == 0 --> REG is NOT the Dest --> R/M REG
		// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
		if d_bit_on {
			builder.WriteString(fmt.Sprintf("%s, %s", reg_str, rm_str))
		} else {
			builder.WriteString(fmt.Sprintf("%s, %s", rm_str, reg_str))
		}
		return
	}

	// Effective Address Calc

	// Check for special Direct Address Case
	if mod_bits == 0b00 && rm_bits == 0b110 {
		i++
		b3 = prog[i]
		i++
		b4 = prog[i]
		disp := uint16(b4) | (uint16(b3) << 8)

		builder.WriteString(fmt.Sprintf("[%d]", disp))
		return
	}

	base_str := rm_effective_map[rm_bits]

	// No Displacement, just end instruction
	if mod_bits == 0b00 {
		rm_str = fmt.Sprintf("[%s]", base_str)
	} else if mod_bits == 0b01 { // 8-bit Displacement
		i++
		b3 = prog[i]
		disp := uint8(b3)

		// The only way to use bp without an offset is to have a zero offset, b/c of the special
		// direct address case above
		if disp == 0 {
			rm_str = fmt.Sprintf("[%s]", base_str)
		} else {
			rm_str = fmt.Sprintf("[%s + %d]", base_str, disp)
		}
	} else if mod_bits == 0b10 { // 16-bit Displacement
		i++
		b3 = prog[i]
		i++
		b4 = prog[i]
		disp := uint16(b3) | (uint16(b4) << 8)

		if disp == 0 {
			rm_str = fmt.Sprintf("[%s]", base_str)
		} else {
			rm_str = fmt.Sprintf("[%s + %d]", base_str, disp)
		}
	}

	// d_bit == 0 --> REG is NOT the Dest --> R/M REG
	// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
	if d_bit_on {
		builder.WriteString(fmt.Sprintf("%s, %s", reg_str, rm_str))
	} else {
		builder.WriteString(fmt.Sprintf("%s, %s", rm_str, reg_str))
	}
}

// MOV 3
func movImmediateToReg() {
	immediateToReg("mov")
}

// TODO: fix variable naming conventions, go snake?
func immediateToReg(opCode string) {
	builder.WriteString(fmt.Sprintf("%s ", opCode))

	w_bit_on := b&0b00001000 != 0
	reg_bits := b & 0b00000111

	var reg_str string
	if w_bit_on {
		reg_str = w_on_map[reg_bits]
	} else {
		reg_str = w_off_map[reg_bits]
	}

	builder.WriteString(fmt.Sprintf("%s, ", reg_str))

	i++
	b2 = prog[i]

	if w_bit_on {
		i++
		b3 = prog[i]

		value := uint16(b2) | (uint16(b3) << 8)

		builder.WriteString(fmt.Sprintf("%d", value))
	} else {
		builder.WriteString(fmt.Sprintf("%d", b2))
	}
}

/*
	Arithmetic
		--> immediate to register/mem == 0b100000XX in first byte
			--> "REG" spot on 2nd byte differentiates the op
			--> see: p. 165-6 of manual

		--> reg/mem and register == 0b00XXX0XX
			--> XXX in the middle differentiates the op

		--> immediate from accumulator == 0b00XXX1XX
			--> XXX in the middle differentiates the op
*/
/*
	--> immediate to register/mem == 0b100000XX in first byte
	--> "REG" spot on 2nd byte differentiates the op
	--> see: p. 165-6 of manual
*/
func arithmeticImmediateToReg() {
	i++
	b2 = prog[i]

	op_bits := (b & 0b00111000) >> 3

	// TODO dupes all over the place!
	arithOpCodes := map[byte]string{
		0b000: "add",
		0b101: "sub",
		0b111: "cmp",
	}

	op_code := arithOpCodes[op_bits]

	immediateToReg(op_code)
}

// --> reg/mem and register == 0b00XXX0XX
// --> XXX in the middle differentiates the op
func arithmeticRegToFromRegMem() {
	op_bits := (b & 0b00111000) >> 3

	// TODO: move to top?? idk...
	arithOpCodes := map[byte]string{
		0b000: "add",
		0b101: "sub",
		0b111: "cmp",
	}

	op_code := arithOpCodes[op_bits]

	regMemToFromEither(op_code)
}

// TODO: don't have the MOV version of this...
func arithmeticImmediateFromAccum() {

}
