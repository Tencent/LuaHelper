package annotateast

// AnnotateFragment 一个片段的注释包含的所有结构, 一个代码注释片段包含多行，每一行大致是一个结构
type AnnotateFragment struct {
	Stats []AnnotateState // 多个子state
	Lines []int           // 每个state的行号
}
