package annotateast

// FieldScopeType filed成员的属性类型：public、protected、private
type FieldScopeType uint8

const (
	_ FieldScopeType = iota

	// FieldScopePublic public属性
	FieldScopePublic = 0

	// FieldScopeProtected protected属性
	FieldScopeProtected = 1

	// FieldScopePrivate private属性
	FieldScopePrivate = 2
)

// FieldColonType filed成员的属性类型 是否为 ： 的成员，默认不会
type FieldColonType uint8

const (
	_ FieldColonType = iota

	// FieldColonNo 默认不为 ：的成员
	FieldColonNo = 0

	// FieldColonYes 为 ：的成员, 例如为下面的
	// ---@field FuncA : fun() : void
	FieldColonYes = 1

	// FieldColonHide 为隐藏 ：的成员, 例如为下面的
	// ---@field AAA
	// ---@field FuncA fun(self:A) : void
	FieldColonHide = 2
)

// EnumType 枚举类型start或end
type EnumType uint8

const (
	_ EnumType = iota

	// EnumTypeNone 非枚举类型
	EnumTypeNone = 0

	// EnumTypeStart 枚举类型的开始
	EnumTypeStart = 1

	// EnumTypeEnd 枚举类型的结束
	EnumTypeEnd = 2
)
