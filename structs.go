package main

type OpType uint8

const (
	OP_MOV OpType = iota
	OP_ADD
	OP_SUB
	OP_CMP
	OP_JNE
	OP_JE
	OP_JL
	OP_JLE
	OP_JB
	OP_JBE
	OP_JP
	OP_JO
	OP_JS
	OP_JNL
	OP_JNLE
	OP_JNB
	OP_JNBE
	OP_JNP
	OP_JNO
	OP_JNS
	OP_LOOP
	OP_LOOPZ
	OP_LOOPNZ
	OP_JCXZ
)

type OperandType uint8

const (
	REGISTER OperandType = iota
	EFFECTIVE_ADDRESS
	S_IMMEDIATE
	U_IMMEDIATE
)			

type Register uint8

const (
	
	// 8-bit
	REG_AL Register = iota
	REG_CL
	REG_DL
	REG_BL
	REG_AH
	REG_CH
	REG_DH
	REG_BH
	// 16-bit
	REG_AX
	REG_CX
	REG_DX
	REG_BX
	REG_SP
	REG_BP
	REG_SI
	REG_DI
)

type EffectiveAddressBase uint8

const (
	EFFECTIVE_ADDRESS_BX_SI EffectiveAddressBase = iota
	EFFECTIVE_ADDRESS_BX_DI
	EFFECTIVE_ADDRESS_BP_SI
	EFFECTIVE_ADDRESS_BP_DI
	EFFECTIVE_ADDRESS_SI
	EFFECTIVE_ADDRESS_DI
	EFFECTIVE_ADDRESS_BP
	EFFECTIVE_ADDRESS_BX
)

type EffectiveAddress struct {
	effectiveAddressBase EffectiveAddressBase
	offset uint16
}

type Operand struct {
	operandType OperandType

	register Register
	effectiveAddress EffectiveAddress
	s_immediate int16
	u_immediate uint16
}

// TODO: need some way to represent that I want to specify byte/word in the (dis-)asm
type Instruction struct {
	operation OpType

	operand_1 Operand
	operand_2 Operand
}
