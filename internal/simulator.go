package internal

import (
	"fmt"
	"strings"

	. "github.com/carvermichael/Sim8086/internal/model"
)

// TODO: move this and its methods into a new file
type Registers struct {
	prev_reg	[8]uint16
	reg		[8]uint16
}

func (reg *Registers) copyToPrev() {
	for i = 0; i < 8; i++ {
		reg.prev_reg[i] = reg.reg[i]
	}
}

func (reg *Registers) getDiff() string {
	b := strings.Builder{}

	for i = 0; i < 8; i++ {
		if reg.prev_reg[i] != reg.reg[i] {
			b.WriteString(fmt.Sprintf("%s: %X --> %X\n", registers_str[i], reg.prev_reg[i], reg.reg[i]))
		}
	}

	return b.String()
}

// assume that, if 8-bit, the data is in the low bits
// TODO: write actual documentation for these, such that you can shift-K them easily --> see what the standards are for go
func (reg *Registers) put(reg_name Register, data uint16) {

	// reg_name constants map directly to 16-bit slots in Registers array
	if reg_name <= REG_DI {
		reg.reg[reg_name] = data
		return
	}

	// if we're dealing with 8-bit registers
	if reg_name > REG_DI {
		if reg_name < REG_AH { // data should only be in low bits
			if data & 0xFF00 > 0 {
				panic("Register put: passed low 8-bit register, but with data in high bits")
			}
		} else { // data should only be in high bits
			if data & 0x00FF > 0 {
				panic("Register put: passed high 8-bit register, but with data in low bits")
			}
		}
	}

	// handle 8-bit cases
	switch reg_name {
	case REG_AL:
		reg.reg[REG_AX] = putLowBits(reg.reg[REG_AX], uint8(data))
	case REG_BL:
		reg.reg[REG_BX] = putLowBits(reg.reg[REG_BX], uint8(data))
	case REG_CL:
		reg.reg[REG_CX] = putLowBits(reg.reg[REG_CX], uint8(data))
	case REG_DL:
		reg.reg[REG_DX] = putLowBits(reg.reg[REG_DX], uint8(data))
	case REG_AH:
		reg.reg[REG_AX] = putHighBits(reg.reg[REG_AX], data)
	case REG_BH:
		reg.reg[REG_BX] = putHighBits(reg.reg[REG_BX], data)
	case REG_CH:
		reg.reg[REG_CX] = putHighBits(reg.reg[REG_CX], data)
	case REG_DH:
		reg.reg[REG_DX] = putHighBits(reg.reg[REG_DX], data)
	}

	return
}

func putLowBits(reg_data uint16, input_data uint8) uint16 {
	reg_data = reg_data & 0xFF00
	return reg_data | uint16(input_data) 
}

func putHighBits(reg_data uint16, input_data uint16) uint16 {
	reg_data = reg_data & 0x00FF
	return uint16(input_data) | reg_data
}

// Note: always returns 8-bit values in low byte of uint16 return value
func (reg *Registers) get(reg_name Register) uint16 {
	
	// reg_name constants map directly to 16-bit slots in Registers array
	if reg_name <= REG_DI {
		return reg.reg[reg_name]
	}

	// handle 8-bit cases
	switch reg_name {
	case REG_AL:
		return reg.reg[REG_AX] & 0x00FF
	case REG_BL:
		return reg.reg[REG_BX] & 0x00FF
	case REG_CL:
		return reg.reg[REG_CX] & 0x00FF
	case REG_DL:
		return reg.reg[REG_DX] & 0x00FF
	case REG_AH:
		return reg.reg[REG_AX] >> 8
	case REG_BH:
		return reg.reg[REG_BX] >> 8
	case REG_CH:
		return reg.reg[REG_CX] >> 8
	case REG_DH:
		return reg.reg[REG_DX] >> 8
	default:
		panic(fmt.Sprintf("Register get called with invalid reg_name, %d", reg_name))
	}
}

func simMOV(reg *Registers, instruction Instruction) {

	dest	:= instruction.Operands[0]
	source	:= instruction.Operands[1]

	// get value
	var data uint16
	switch source.OperandType {
	case U_IMMEDIATE, S_IMMEDIATE:
		if source.Wide {
			data = (uint16(source.Immediate_High) << 8) | uint16(source.Immediate_Low)
		} else {
			data = uint16(source.Immediate_Low)
		}
	case REGISTER:
		data = reg.get(source.Register)
	default:
		panic("MOV, attempt to simulate unimplemented data source")
	}

	// get destination
	switch dest.OperandType {
	case REGISTER:
		reg.put(instruction.Operands[0].Register, data)
	default:
		panic("MOV, attempt to simulate unimplemented destination")
	}
}

// TODO: don't like passing this string builder around here...
func dispatchInstruction(reg *Registers, instruction Instruction, b *strings.Builder) {
	reg.copyToPrev()
	GetInstructionString(instruction, b)

	switch instruction.Operation {
	case OP_MOV:
		simMOV(reg, instruction)
	default: 
		return
	}

	b.WriteString("// ; ")
	b.WriteString(reg.getDiff())
}

func SimulateInstructions(instructions []Instruction) {
	b := strings.Builder{}

	reg := Registers{}

	for _, v := range instructions {
		dispatchInstruction(&reg, v, &b)		
	}

	fmt.Print(b.String())
}


