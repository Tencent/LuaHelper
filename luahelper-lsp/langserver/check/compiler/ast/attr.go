package ast

// LocalAttr  local var attribute
type LocalAttr uint8 // attribute type for local varible

const (
	// VDKREG no attr
	VDKREG LocalAttr = 0
	// RDKTOCLOSE close attr
	RDKTOCLOSE LocalAttr = 1
	// RDKCONST const attr
	RDKCONST LocalAttr = 2
)
