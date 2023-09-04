package internal

import (
	"io/ioutil"

	. "github.com/carvermichael/Sim8086/internal/model"
)

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

func GetASMFromFile(fileName string) []Instruction {

	input, err = ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	instructions := make([]Instruction, 0)
	for i = 0; i < len(input); {
		instructions = append(instructions, getNextInstruction())
	}

	return instructions
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

	return instruction
}

// MOV #1
func movRegToFromRegMem() Instruction {
	return regMemToFromEither(OP_MOV)
}

func regMemToFromEither(opType OpType) Instruction {
	instruction := Instruction{}

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

	reg_operand := getReg(reg_bits, w_bit_on)

	rm_operand := getRegMemOperand(mod_bits, rm_bits, w_bit_on)

	// d_bit == 0 --> REG is NOT the Dest --> R/M REG
	// d_bit == 1 --> REG IS the Dest 	  --> REG R/M
	if d_bit_on {
		instruction.Operands = []Operand{
			reg_operand,
			rm_operand,
		}
	} else {
		instruction.Operands = []Operand{
			rm_operand,
			reg_operand,
		}
	}

	return instruction
}

func getReg(reg_bits byte, w_bit_on bool) Operand {
	reg_operand := Operand{
		OperandType: REGISTER,
	}

	if w_bit_on {
		reg_operand.Register = w_on_arr_enum[reg_bits]
	} else {
		reg_operand.Register = w_off_arr_enum[reg_bits]
	}
	return reg_operand
}

// r/m field --> Effective Address Calc
func getRegMemOperand(mod_bits byte, rm_bits byte, w_bit_on bool) Operand {

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

		return rm_operand
	}

	rm_operand.OperandType = EFFECTIVE_ADDRESS
	rm_operand.EffectiveAddress = EffectiveAddress{
		EffectiveAddressBase: rm_effective_arr_enum[rm_bits],
	}

	// No Displacement, just end instruction
	// if mod_bits == 0b00 {
	// 	rm_str = fmt.Sprintf("[%s]", base_str)
	if mod_bits == 0b01 { // 8-bit Displacement
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		disp := uint8(b3)
		rm_operand.EffectiveAddress.Offset = uint16(disp)

	} else if mod_bits == 0b10 { // 16-bit Displacement
		//i++
		//b3 = prog[i]
		loadByteNum(3)
		//i++
		//b4 = prog[i]
		loadByteNum(4)
		disp := uint16(b3) | (uint16(b4) << 8)

		rm_operand.EffectiveAddress.Offset = uint16(disp)
	}
	return rm_operand
}


// MOV 3
func movImmediateToReg() Instruction {
	instruction := Instruction{}
	instruction.Operation = OP_MOV

	w_bit_on := b&0b00001000 != 0
	reg_bits := b&0b00000111

	operand1 := getReg(reg_bits, w_bit_on)

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

		operand2.Immediate_High = b3
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

	op_enum := arithOpCodes_enum[op_bits]
	instruction.Operation = op_enum

	operand1 := getRegMemOperand(mod_bits, rm_bits, w_bit_on)

	// BUG: Getting the 5th byte here will be incorrect when getRegMemOperand doesn't get both the 3rd and 4th bytes.
	// TODO: move to passing a pointer around to know what byte you're at.
	loadByteNum(5)
	
	instruction.Wide = w_bit_on
	operand2 := Operand{}

	// instruction version

	if !s_bit_on && w_bit_on {
		loadByteNum(6)
	}

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

	opType := arithOpCodes_enum[op_bits]

	return regMemToFromEither(opType)
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

	instruction.Operation = arithOpCodes_enum[op_bits]

	loadByteNum(2)

	w_bit_on := b&0b00000001 != 0

	operand2.Immediate_Low = b2

	instruction.Wide = w_bit_on
	if w_bit_on {
		loadByteNum(3)

		operand2.Immediate_High = b3
	}

	instruction.Operands = []Operand{
		operand1, operand2,
	}

	return instruction
}

func jumps() Instruction {
	var instruction Instruction

	switch b {
	// jne/jnz
	case 0b01110101:
		instruction = Instruction{Operation: OP_JNE}
		break
	// je/jz
	case 0b01110100:
		instruction = Instruction{Operation: OP_JE}
		break
	// jl/jnge
	case 0b01111100:
		instruction = Instruction{Operation: OP_JL}
		break
	// jle/jng
	case 0b01111110:
		instruction = Instruction{Operation: OP_JLE}
		break
	// jb/jnae
	case 0b01110010:
		instruction = Instruction{Operation: OP_JB}
		break
	// jbe/jna
	case 0b01110110:
		instruction = Instruction{Operation: OP_JBE}
		break
	// jp/jpe
	case 0b01111010:
		instruction = Instruction{Operation: OP_JP}
		break
	// jo
	case 0b01110000:
		instruction = Instruction{Operation: OP_JO}
		break
	// js
	case 0b01111000:
		instruction = Instruction{Operation: OP_JS}
		break
	// jnl/jge
	case 0b01111101:
		instruction = Instruction{Operation: OP_JNL}
		break
	// jnle/jg
	case 0b01111111:
		instruction = Instruction{Operation: OP_JNLE}
		break
	// jnb/jae
	case 0b01110011:
		instruction = Instruction{Operation: OP_JNB}
		break
	// jnbe/ja
	case 0b01110111:
		instruction = Instruction{Operation: OP_JNBE}
		break
	// jnp/jpo
	case 0b01111011:
		instruction = Instruction{Operation: OP_JNP}
		break
	// jno
	case 0b01110001:
		instruction = Instruction{Operation: OP_JNO}
		break
	// jns
	case 0b01111001:
		instruction = Instruction{Operation: OP_JNS}
		break
	// loop
	case 0b11100010:
		instruction = Instruction{Operation: OP_LOOP}
		break
	// loopz/loope
	case 0b11100001:
		instruction = Instruction{Operation: OP_LOOPZ}
		break
	// loopnz/loopne
	case 0b11100000:
		instruction = Instruction{Operation: OP_LOOPNZ}
		break
	// jcxz
	case 0b11100011:
		instruction = Instruction{Operation: OP_JCXZ}
		break
	default:
		instruction = Instruction{Operation: OP_NONE}
		return instruction
	}

	loadByteNum(2)

	// TODO: don't like this at all, figure out better place for jump data
	operand := Operand{
		OperandType: S_IMMEDIATE,
		Immediate_Low: b2,
	}
	instruction.Operands = []Operand{
		operand,
	}

	return instruction
}
