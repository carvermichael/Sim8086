package internal

import (
	"strings"
	. "github.com/carvermichael/Sim8086/internal/model"
	"io/ioutil"
	"fmt"
)

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

var input []byte
var i int                           // current index int = 0
var b, b2, b3, b4, b5, b6 byte      // current bytes (TODO: array? --> pointers?)
var disassemBuilder strings.Builder // builds disassembly output
var instructions []Instruction      // Intermediate Representation

var err error

// TODO: this is probably really dumb, could probably use the same byte for everything (and a pointer to the byte at that)
func loadByteNum(number int) {
	switch number {
	case 1:
		b = input[i]
	case 2:
		i++
		b2 = input[i]
	case 3:
		i++
		b3 = input[i]
	case 4:
		i++
		b4 = input[i]
	case 5:
		i++
		b5 = input[i]
	case 6:
		i++
		b6 = input[i]
	}
}

func GetASMFromFile(fileName string) string {

	input, err = ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	for i = 0; i < len(input); {
		//b = prog[i]
		loadByteNum(1)

		if (b & 0b11111100) == 0b10001000 {
			movRegToFromRegMem()
		} else if (b & 0b11110000) == 0b10110000 {
			movImmediateToReg()
		} else if (b & 0b11111100) == 0b10000000 {
			// immediate to register/mem == 0b100000XX in first byte
			arithmeticImmediateToRegMem()
		} else if (b & 0b11000100) == 0b00000000 {
			// --> reg/mem and register == 0b00XXX0XX
			arithmeticRegToFromRegMem()
		} else if (b & 0b11000110) == 0b00000100 {
			// --> immediate from accumulator == 0b00XXX10X
			arithmeticImmediateToAccum()
		} else {
			jumps()
		}

		i++

		disassemBuilder.WriteString("\n")
	}

	return disassemBuilder.String()
}

// MOV #1
func movRegToFromRegMem() {
	regMemToFromEither("mov")
}

func regMemToFromEither(opCode string) {
	disassemBuilder.WriteString(fmt.Sprintf("%s ", opCode))

	// get d, w, mod, reg, and r/m first
	w_bit_on := (b & 0b00000001) != 0
	d_bit_on := (b & 0b00000010) != 0

	//i++
	//b2 = prog[i]
	loadByteNum(2)

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

	rm_str := getEffectiveAddress(mod_bits, rm_bits, w_bit_on)

	// d_bit == 0 --> REG is NOT the Dest --> R/M REG
	// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
	if d_bit_on {
		disassemBuilder.WriteString(fmt.Sprintf("%s, %s", reg_str, rm_str))
	} else {
		disassemBuilder.WriteString(fmt.Sprintf("%s, %s", rm_str, reg_str))
	}
}

// r/m field --> Effective Address Calc

// TODO: rename --> get r/m string
func getEffectiveAddress(mod_bits byte, rm_bits byte, w_bit_on bool) string {

	var rm_str string

	// Register Mode (Register to Register)
	if mod_bits == 0b11 {
		if w_bit_on {
			rm_str = w_on_map[rm_bits]
		} else {
			rm_str = w_off_map[rm_bits]
		}

		return rm_str
	}

	//TODO: this is bugged AND I don't think any of the test files use this case... /shrug
	// Check for special Direct Address Case
	if mod_bits == 0b00 && rm_bits == 0b110 {
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		//i++
		//b4 = prog[i]
		loadByteNum(4)
		disp := uint16(b3) | (uint16(b4) << 8)

		return fmt.Sprintf("[%d]", disp)
	}

	base_str := rm_effective_map[rm_bits]

	// No Displacement, just end instruction
	if mod_bits == 0b00 {
		rm_str = fmt.Sprintf("[%s]", base_str)
	} else if mod_bits == 0b01 { // 8-bit Displacement
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		disp := uint8(b3)

		// The only way to use bp without an offset is to have a zero offset, b/c of the special
		// direct address case above
		if disp == 0 {
			rm_str = fmt.Sprintf("[%s]", base_str)
		} else {
			rm_str = fmt.Sprintf("[%s + %d]", base_str, disp)
		}
	} else if mod_bits == 0b10 { // 16-bit Displacement
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		//i++
		//b4 = prog[i]
		loadByteNum(4)
		disp := uint16(b3) | (uint16(b4) << 8)

		if disp == 0 {
			rm_str = fmt.Sprintf("[%s]", base_str)
		} else {
			rm_str = fmt.Sprintf("[%s + %d]", base_str, disp)
		}
	}
	return rm_str
}

// MOV 3
func movImmediateToReg() {
	immediateToReg("mov")
}

// TODO: fix variable naming conventions, go snake?
func immediateToReg(opCode string) {
	disassemBuilder.WriteString(fmt.Sprintf("%s ", opCode))

	w_bit_on := b&0b00001000 != 0
	reg_bits := b & 0b00000111

	var reg_str string
	if w_bit_on {
		reg_str = w_on_map[reg_bits]
	} else {
		reg_str = w_off_map[reg_bits]
	}

	disassemBuilder.WriteString(fmt.Sprintf("%s, ", reg_str))

	//i++
	//b2 = prog[i]
	loadByteNum(2)

	if w_bit_on {
		//i++
		//b3 = prog[i]
		loadByteNum(3)

		value := uint16(b2) | (uint16(b3) << 8)

		disassemBuilder.WriteString(fmt.Sprintf("%d", value))
	} else {
		disassemBuilder.WriteString(fmt.Sprintf("%d", b2))
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
// TODO
func arithmeticImmediateToRegMem() {
	// op_code rm_str [data]

	s_bit_on := b&0b00000010 != 0
	w_bit_on := b&0b00000001 != 0

	loadByteNum(2)

	op_bits := (b2 & 0b00111000) >> 3

	mod_bits := b2 & 0b11000000 >> 6
	rm_bits := b2 & 0b00000111

	// TODO dupes all over the place! --> also, this could also just be a switch (map is computational overkill for suresies)
	arithOpCodes := map[byte]string{
		0b000: "add",
		0b101: "sub",
		0b111: "cmp",
	}

	op_code := arithOpCodes[op_bits]
	disassemBuilder.WriteString(fmt.Sprintf("%s ", op_code))

	rm_str := getEffectiveAddress(mod_bits, rm_bits, w_bit_on)

	loadByteNum(5)

	// TODO: gotta be a better way to write this...
	if w_bit_on {
		size_str := "word"
		if !s_bit_on {
			loadByteNum(6)
			data := uint16(b5) | (uint16(b6) << 8)
			// Only when in register mode, we don't need to specify byte/word, b/c that's implicit in the register name.
			if mod_bits == 0b11 {
				disassemBuilder.WriteString(fmt.Sprintf("%s %d", rm_str, data))
			} else {
				disassemBuilder.WriteString(fmt.Sprintf("%s %s %d", size_str, rm_str, data))
			}
		} else {
			data := int16(b5)
			if mod_bits == 0b11 {
				disassemBuilder.WriteString(fmt.Sprintf("%s %d", rm_str, data))
			} else {
				disassemBuilder.WriteString(fmt.Sprintf("%s %s %d", size_str, rm_str, data))
			}
		}
	} else {
		size_str := "byte"
		if !s_bit_on {
			data := int8(b5)
			if mod_bits == 0b11 {
				disassemBuilder.WriteString(fmt.Sprintf("%s %d", rm_str, data))
			} else {
				disassemBuilder.WriteString(fmt.Sprintf("%s %s %d", size_str, rm_str, data))
			}
		} else {
			data := uint8(b5)
			if mod_bits == 0b11 {
				disassemBuilder.WriteString(fmt.Sprintf("%s %d", rm_str, data))
			} else {
				disassemBuilder.WriteString(fmt.Sprintf("%s %s %d", size_str, rm_str, data))
			}
		}
	}

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
func arithmeticImmediateToAccum() {
	op_bits := (b & 0b00111000) >> 3

	// TODO dupes all over the place! --> also, this could also just be a switch (map is computational overkill for suresies)
	arithOpCodes := map[byte]string{
		0b000: "add",
		0b101: "sub",
		0b111: "cmp",
	}

	op_code := arithOpCodes[op_bits]
	disassemBuilder.WriteString(fmt.Sprintf("%s ax ", op_code))

	loadByteNum(2)

	w_bit_on := b&0b00000001 != 0
	if w_bit_on {
		loadByteNum(3)

		data := uint16(b2) | (uint16(b3) << 8)
		disassemBuilder.WriteString(fmt.Sprintf("%d", data))
	} else {
		data := b2
		disassemBuilder.WriteString(fmt.Sprintf("%d", data))
	}

}

func jumps() {
	var jmp_str string

	switch b {
	// jne/jnz
	case 0b01110101:
		jmp_str = "jne"
		break
	// je/jz
	case 0b01110100:
		jmp_str = "je"
		break
	// jl/jnge
	case 0b01111100:
		jmp_str = "jl"
		break
	// jle/jng
	case 0b01111110:
		jmp_str = "jle"
		break
	// jb/jnae
	case 0b01110010:
		jmp_str = "jb"
		break
	// jbe/jna
	case 0b01110110:
		jmp_str = "jbe"
		break
	// jp/jpe
	case 0b01111010:
		jmp_str = "jp"
		break
	// jo
	case 0b01110000:
		jmp_str = "jo"
		break
	// js
	case 0b01111000:
		jmp_str = "js"
		break
	// jnl/jge
	case 0b01111101:
		jmp_str = "jnl"
		break
	// jnle/jg
	case 0b01111111:
		jmp_str = "jnle"
		break
	// jnb/jae
	case 0b01110011:
		jmp_str = "jnb"
		break
	// jnbe/ja
	case 0b01110111:
		jmp_str = "jnbe"
		break
	// jnp/jpo
	case 0b01111011:
		jmp_str = "jnp"
		break
	// jno
	case 0b01110001:
		jmp_str = "jno"
		break
	// jns
	case 0b01111001:
		jmp_str = "jns"
		break
	// loop
	case 0b11100010:
		jmp_str = "loop"
		break
	// loopz/loope
	case 0b11100001:
		jmp_str = "loopz"
		break
	// loopnz/loopne
	case 0b11100000:
		jmp_str = "loopnz"
		break
	// jcxz
	case 0b11100011:
		jmp_str = "jcxz"
		break
	default:
		jmp_str = "uhhhhhh"
	}

	loadByteNum(2)
	data := int8(b2)

	disassemBuilder.WriteString(fmt.Sprintf("%s ... ; %d", jmp_str, data))
}
