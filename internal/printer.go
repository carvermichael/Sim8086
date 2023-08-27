package internal

import (
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

func PrintInstructions(instructions []Instruction) string {

	var b strings.Builder

	for _, v := range instructions {
		printInstruction(v, &b)
	}

	return b.String()
}

func printInstruction(instruction Instruction, b *strings.Builder) {

	// Op Mnemoic
	(*b).WriteString(opCodes[instruction.Operation])

	for _, v := range instruction.Operands {
		printOperand(v, b)
	}

	// End line
	(*b).WriteString("\n")
}

func printOperand(operand Operand, b *strings.Builder) {
	
	switch operand.OperandType {
	case REGISTER:
		(*b).WriteString(registers[operand.Register])
		break
	case EFFECTIVE_ADDRESS:
		break
	case DIRECT_ADDRESS:
		break
	case S_IMMEDIATE:
		break
	case U_IMMEDIATE:
		break
	}

	return
}

