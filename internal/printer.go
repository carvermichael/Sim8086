package internal

import (
	"fmt"
	"strings"

	. "github.com/carvermichael/Sim8086/internal/model"
)

// TODO: gotta keep this in line with the OP code enum in the model file, not wild about it, maybe generate it every time?
var opCodes = []string{ 
	"none",
	"mov",
	"add",
	"sub",
	"cmp",
	"jne",
	"je",
	"jl",
	"jle",
	"jb",
	"jbe",
	"jp",
	"jo",
	"js",
	"jnl",
	"jnle",
	"jnb",
	"jnbe",
	"jnp",
	"jno",
	"jns",
	"loop",
	"loopz",
	"loopnz",
	"jcxz",
}

var registers = []string{
	"al",
	"cl",
	"dl",
	"bl",
	"ah",
	"ch",
	"dh",
	"bh",
	"ax",
	"cx",
	"dx",
	"bx",
	"sp",
	"bp",
	"si",
	"di",
}

var effective_adds = []string{
	"bx + si",
	"bx + di",
	"bp + si",
	"bp + di",
	"si",
	"di",
	"bp",
	"bx",
}

func PrintInstructions(instructions []Instruction) string {

	var b strings.Builder

	for _, v := range instructions {
		printInstruction(v, &b)
	}

	return b.String()
}

func printInstruction(instruction Instruction, b *strings.Builder) {

	// Op Mnemoic
	(*b).WriteString(opCodes[instruction.Operation] + " ")
	
	if instruction.Specify_Size {
		if instruction.Wide {
			(*b).WriteString("word ")
		} else {
			(*b).WriteString("byte ")
		}
	}

	for _, v := range instruction.Operands {
		printOperand(v, instruction.Wide, b)
		(*b).WriteString(" ")
	}

	// End line
	(*b).WriteString("\n")
}

func printOperand(operand Operand, wide bool, b *strings.Builder) {
	
	switch operand.OperandType {
	case REGISTER:
		(*b).WriteString(registers[operand.Register])
		break
	case EFFECTIVE_ADDRESS:
		base := operand.EffectiveAddress.EffectiveAddressBase
		offset := operand.EffectiveAddress.Offset

		if offset == 0 {
			(*b).WriteString(fmt.Sprintf("[%s]", effective_adds[base]))
		} else {
			(*b).WriteString(fmt.Sprintf("[%s + %d]", effective_adds[base], offset))
		}
		break
	case DIRECT_ADDRESS:
		(*b).WriteString(fmt.Sprintf("[%d]", operand.DirectAddress))
		break
	case S_IMMEDIATE:
		if wide {
			value := int16(int16(operand.Immediate_Low) | (int16(operand.Immediate_High) << 8))
			(*b).WriteString(fmt.Sprintf("%d", value))
		} else {
			// TODO: apparently this never occurs in 0041
			value := int8(operand.Immediate_Low)
			(*b).WriteString(fmt.Sprintf("%d", value))
		}
		break
	case U_IMMEDIATE:
		if wide {
			value := uint16(uint16(operand.Immediate_Low) | (uint16(operand.Immediate_High) << 8))
			(*b).WriteString(fmt.Sprintf("%d", value))
		} else {
			value := operand.Immediate_Low
			(*b).WriteString(fmt.Sprintf("%d", value))
		}
		break
	}

	return
}

