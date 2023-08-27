package model

type OpType uint8

const (
	OP_NONE OpType = iota // OP_NONE necessary for blank spots in arr lookup -- see: arithOpCodes && jump switch default (unknown opCode)
	OP_MOV 
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
	DIRECT_ADDRESS
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
	EffectiveAddressBase EffectiveAddressBase
	Offset               uint16
}

type Operand struct {
	OperandType		OperandType

	Register		Register
	EffectiveAddress	EffectiveAddress
	DirectAddress		uint16
	Immediate_Low		byte
	Immediate_High		byte
}

type Instruction struct {
	Operation OpType

	Operands  []Operand

	Wide		bool
	Specify_Size	bool
}
