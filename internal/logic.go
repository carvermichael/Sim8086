package internal

import (
	"fmt"
	"io/ioutil"
	"strings"

	. "github.com/carvermichael/Sim8086/internal/model"
)

var w_on_arr_str = []string{
	"ax",
	"cx",
	"dx",
	"bx",
	"sp",
	"bp",
	"si",
	"di",
}

var w_off_arr_str = []string{
	"al",
	"cl",
	"dl",
	"bl",
	"ah",
	"ch",
	"dh",
	"bh",
}

var rm_effective_arr_str = []string{
	"bx + si",
	"bx + di",
	"bp + si",
	"bp + di",
	"si",
	"di",
	"bp",
	"bx",
}


var w_on_arr_enum = []Register{
	REG_AX,
	REG_CX,
	REG_DX,
	REG_BX,
	REG_SP,
	REG_BP,
	REG_SI,
	REG_DI,
}

var w_off_arr_enum = []Register{
	REG_AL,
	REG_CL,
	REG_DL,
	REG_BL,
	REG_AH,
	REG_CH,
	REG_DH,
	REG_BH,
}

var rm_effective_arr_enum = []EffectiveAddressBase{
	EFFECTIVE_ADDRESS_BX_SI,
	EFFECTIVE_ADDRESS_BX_DI,
	EFFECTIVE_ADDRESS_BP_SI,
	EFFECTIVE_ADDRESS_BP_DI,
	EFFECTIVE_ADDRESS_SI,
	EFFECTIVE_ADDRESS_DI,
	EFFECTIVE_ADDRESS_BP,
	EFFECTIVE_ADDRESS_BX,
}

var arithOpCodes_str = []string{
	"add",
	"",
	"",
	"",
	"",
	"sub",
	"",
	"cmp",
}

var arithOpCodes_enum = []OpType{
	OP_ADD,
	OP_NONE,
	OP_NONE,
	OP_NONE,
	OP_NONE,
	OP_SUB,
	OP_NONE,
	OP_CMP,
}

var input []byte
var i int                           // current index int = 0
var b, b2, b3, b4, b5, b6 byte      // current bytes (TODO: array? --> pointers?)
var disassemBuilder strings.Builder // builds disassembly output

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

func GetASMFromFile(fileName string) (string, []Instruction) {

	input, err = ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	instructions := make([]Instruction, 0)
	for i = 0; i < len(input); {
		instructions = append(instructions, getNextInstruction())
	}

	return disassemBuilder.String(), instructions
}

func getNextInstruction() Instruction {
	
	//b = prog[i]
	loadByteNum(1)

	var instruction Instruction

	if (b & 0b11111100) == 0b10001000 {
		instruction = movRegToFromRegMem()
	} else if (b & 0b11110000) == 0b10110000 {
		instruction = movImmediateToReg()
	} else if (b & 0b11111100) == 0b10000000 {
		// immediate to register/mem == 0b100000XX in first byte
		instruction = arithmeticImmediateToRegMem()
	} else if (b & 0b11000100) == 0b00000000 {
		// --> reg/mem and register == 0b00XXX0XX
		instruction = arithmeticRegToFromRegMem()
	} else if (b & 0b11000110) == 0b00000100 {
		// --> immediate from accumulator == 0b00XXX10X
		instruction = arithmeticImmediateToAccum()
	} else {
		instruction = jumps()
	}

	i++

	disassemBuilder.WriteString("\n")

	return instruction
}

// MOV #1
func movRegToFromRegMem() Instruction {
	return regMemToFromEither("mov", OP_MOV)
}

func regMemToFromEither(opCode string, opType OpType) Instruction {
	instruction := Instruction{}

	disassemBuilder.WriteString(fmt.Sprintf("%s ", opCode))
	instruction.Operation = opType

	// get d, w, mod, reg, and r/m first
	w_bit_on := (b & 0b00000001) != 0
	instruction.Wide = w_bit_on


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

	reg_str, reg_operand := getReg(reg_bits, w_bit_on)

	rm_str, rm_operand := getRegMemOperand(mod_bits, rm_bits, w_bit_on)

	// d_bit == 0 --> REG is NOT the Dest --> R/M REG
	// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
	if d_bit_on {
		disassemBuilder.WriteString(fmt.Sprintf("%s, %s", reg_str, rm_str))
		instruction.Operands = []Operand{
			reg_operand,
			rm_operand,
		}
	} else {
		disassemBuilder.WriteString(fmt.Sprintf("%s, %s", rm_str, reg_str))
		instruction.Operands = []Operand{
			rm_operand,
			reg_operand,
		}
	}

	return instruction
}

func getReg(reg_bits byte, w_bit_on bool) (string, Operand) {
	var reg_str string
	reg_operand := Operand{
		OperandType: REGISTER,
	}

	if w_bit_on {
		reg_str = w_on_arr_str[reg_bits]
		reg_operand.Register = w_on_arr_enum[reg_bits]
	} else {
		reg_str = w_off_arr_str[reg_bits]
		reg_operand.Register = w_off_arr_enum[reg_bits]
	}
	return reg_str, reg_operand
}

// r/m field --> Effective Address Calc
func getRegMemOperand(mod_bits byte, rm_bits byte, w_bit_on bool) (string, Operand) {

	var rm_str string
	rm_operand := Operand{}

	// Register Mode (Register to Register)
	if mod_bits == 0b11 {
		return getReg(rm_bits, w_bit_on)
	}

	// TODO: this is bugged AND I don't think any of the test files use this case... /shrug
	// Check for special Direct Address Case
	if mod_bits == 0b00 && rm_bits == 0b110 {
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		//i++
		//b4 = prog[i]
		loadByteNum(4)
		disp := uint16(b3) | (uint16(b4) << 8)

		rm_operand.OperandType = DIRECT_ADDRESS
		rm_operand.DirectAddress = disp

		return fmt.Sprintf("[%d]", disp), rm_operand
	}

	base_str := rm_effective_arr_str[rm_bits]

	rm_operand.OperandType = EFFECTIVE_ADDRESS
	rm_operand.EffectiveAddress = EffectiveAddress{
		EffectiveAddressBase: rm_effective_arr_enum[rm_bits],
	}

	// No Displacement, just end instruction
	if mod_bits == 0b00 {
		rm_str = fmt.Sprintf("[%s]", base_str)
	} else if mod_bits == 0b01 { // 8-bit Displacement
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		disp := uint8(b3)
		rm_operand.EffectiveAddress.Offset = uint16(disp)

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

		rm_operand.EffectiveAddress.Offset = uint16(disp)

		if disp == 0 {
			rm_str = fmt.Sprintf("[%s]", base_str)
		} else {
			rm_str = fmt.Sprintf("[%s + %d]", base_str, disp)
		}
	}
	return rm_str, rm_operand
}


// MOV 3
func movImmediateToReg() Instruction {
	instruction := Instruction{}
	instruction.Operation = OP_MOV

	disassemBuilder.WriteString(fmt.Sprintf("mov "))

	w_bit_on := b&0b00001000 != 0
	reg_bits := b&0b00000111

	var reg_str string
	reg_str, operand1 := getReg(reg_bits, w_bit_on)

	disassemBuilder.WriteString(fmt.Sprintf("%s, ", reg_str))

	//i++
	//b2 = prog[i]
	loadByteNum(2)

	operand2 := Operand{
		OperandType: U_IMMEDIATE,
		Immediate_Low: b2,
	}

	// TODO: do something like this after printing logic is gone...
	// instruction := Instruction{
	// 	Operation: OP_MOV,
	// 	Operand_1: getReg(reg_bits, w_bit_on),
	// 	Operand_2: Operand{
	// 		OperandType: U_IMMEDIATE,
	// 		Immediate_Low: b2,
	// 	},
	// 	Wide: w_bit_on,
	// }

	if w_bit_on {
		//i++
		//b3 = prog[i]
		loadByteNum(3)

		value := uint16(b2) | (uint16(b3) << 8)

		disassemBuilder.WriteString(fmt.Sprintf("%d", value))
		operand2.Immediate_High = b3
	} else {
		disassemBuilder.WriteString(fmt.Sprintf("%d", b2))
	}

	instruction.Operands = []Operand{
		operand1, operand2,
	}
	
	return instruction
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

// Form: op_code rm_str [data]
func arithmeticImmediateToRegMem() Instruction {
	instruction := Instruction{}

	s_bit_on := b&0b00000010 != 0
	w_bit_on := b&0b00000001 != 0

	loadByteNum(2)

	op_bits := (b2 & 0b00111000) >> 3

	mod_bits := b2 & 0b11000000 >> 6
	rm_bits := b2 & 0b00000111

	op_code := arithOpCodes_str[op_bits]
	disassemBuilder.WriteString(fmt.Sprintf("%s ", op_code))

	op_enum := arithOpCodes_enum[op_bits]
	instruction.Operation = op_enum

	rm_str, operand1 := getRegMemOperand(mod_bits, rm_bits, w_bit_on)

	// BUG: Getting the 5th byte here will be incorrect when getRegMemOperand doesn't get both the 3rd and 4th bytes.
	// TODO: move to passing a pointer around to know what byte you're at.
	loadByteNum(5)
	
	instruction.Wide = w_bit_on
	operand2 := Operand{}

	// printing version -- remove later, of course
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

	// instruction version
	operand2.Immediate_Low = b5
	if w_bit_on {
		operand2.Immediate_High = b6
	}

	if s_bit_on {
		operand2.OperandType = S_IMMEDIATE
	} else {
		operand2.OperandType = U_IMMEDIATE
	}

	if mod_bits != 0b11 {
		instruction.Specify_Size = true
	}
	instruction.Operands = []Operand{
		operand1, operand2,
	}

	return instruction
}

// --> reg/mem and register == 0b00XXX0XX
// --> XXX in the middle differentiates the op
func arithmeticRegToFromRegMem() Instruction {
	op_bits := (b & 0b00111000) >> 3

	op_code := arithOpCodes_str[op_bits]
	opType := arithOpCodes_enum[op_bits]

	return regMemToFromEither(op_code, opType)
}

// TODO: don't have the MOV version of this...
func arithmeticImmediateToAccum() Instruction {
	operand1 := Operand{
		OperandType: REGISTER,
		Register: REG_AX,
	}
	operand2 := Operand{
		OperandType: U_IMMEDIATE,
	}
	instruction := Instruction{}

	op_bits := (b & 0b00111000) >> 3

	op_code := arithOpCodes_str[op_bits]
	instruction.Operation = arithOpCodes_enum[op_bits]

	disassemBuilder.WriteString(fmt.Sprintf("%s ax ", op_code))

	loadByteNum(2)

	w_bit_on := b&0b00000001 != 0

	operand2.Immediate_Low = b2

	instruction.Wide = w_bit_on
	if w_bit_on {
		loadByteNum(3)

		data := uint16(b2) | (uint16(b3) << 8)

		disassemBuilder.WriteString(fmt.Sprintf("%d", data))

		operand2.Immediate_High = b3
	} else {
		data := b2
		disassemBuilder.WriteString(fmt.Sprintf("%d", data))
	}

	instruction.Operands = []Operand{
		operand1, operand2,
	}


	return instruction
}

func jumps() Instruction {
	var jmp_str string
	var instruction Instruction

	switch b {
	// jne/jnz
	case 0b01110101:
		jmp_str = "jne"
		instruction = Instruction{Operation: OP_JNE}
		break
	// je/jz
	case 0b01110100:
		jmp_str = "je"
		instruction = Instruction{Operation: OP_JE}
		break
	// jl/jnge
	case 0b01111100:
		jmp_str = "jl"
		instruction = Instruction{Operation: OP_JL}
		break
	// jle/jng
	case 0b01111110:
		jmp_str = "jle"
		instruction = Instruction{Operation: OP_JLE}
		break
	// jb/jnae
	case 0b01110010:
		jmp_str = "jb"
		instruction = Instruction{Operation: OP_JB}
		break
	// jbe/jna
	case 0b01110110:
		jmp_str = "jbe"
		instruction = Instruction{Operation: OP_JBE}
		break
	// jp/jpe
	case 0b01111010:
		jmp_str = "jp"
		instruction = Instruction{Operation: OP_JP}
		break
	// jo
	case 0b01110000:
		jmp_str = "jo"
		instruction = Instruction{Operation: OP_JO}
		break
	// js
	case 0b01111000:
		jmp_str = "js"
		instruction = Instruction{Operation: OP_JS}
		break
	// jnl/jge
	case 0b01111101:
		jmp_str = "jnl"
		instruction = Instruction{Operation: OP_JNL}
		break
	// jnle/jg
	case 0b01111111:
		jmp_str = "jnle"
		instruction = Instruction{Operation: OP_JNLE}
		break
	// jnb/jae
	case 0b01110011:
		jmp_str = "jnb"
		instruction = Instruction{Operation: OP_JNB}
		break
	// jnbe/ja
	case 0b01110111:
		jmp_str = "jnbe"
		instruction = Instruction{Operation: OP_JNBE}
		break
	// jnp/jpo
	case 0b01111011:
		jmp_str = "jnp"
		instruction = Instruction{Operation: OP_JNP}
		break
	// jno
	case 0b01110001:
		jmp_str = "jno"
		instruction = Instruction{Operation: OP_JNO}
		break
	// jns
	case 0b01111001:
		jmp_str = "jns"
		instruction = Instruction{Operation: OP_JNS}
		break
	// loop
	case 0b11100010:
		jmp_str = "loop"
		instruction = Instruction{Operation: OP_LOOP}
		break
	// loopz/loope
	case 0b11100001:
		jmp_str = "loopz"
		instruction = Instruction{Operation: OP_LOOPZ}
		break
	// loopnz/loopne
	case 0b11100000:
		jmp_str = "loopnz"
		instruction = Instruction{Operation: OP_LOOPNZ}
		break
	// jcxz
	case 0b11100011:
		jmp_str = "jcxz"
		instruction = Instruction{Operation: OP_JCXZ}
		break
	default:
		jmp_str = "uhhhhhh"
		instruction = Instruction{Operation: OP_NONE}
		return instruction
	}

	loadByteNum(2)
	data := int8(b2)

	// TODO: don't like this at all, figure out better place for jump data
	operand := Operand{
		OperandType: S_IMMEDIATE,
		Immediate_Low: b2,
	}
	instruction.Operands = []Operand{
		operand,
	}


	disassemBuilder.WriteString(fmt.Sprintf("%s ... ; %d", jmp_str, data))

	return instruction
}
