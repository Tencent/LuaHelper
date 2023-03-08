package lexer

// Location line and colomn
type Location struct {
	StartLine   int // from 1
	StartColumn int // from 1
	EndLine     int
	EndColumn   int
}

// ParseError check的错误信息
type ParseError struct {
	ErrStr      string   // 简单的错误信息
	Loc         Location // 错误的位置区域
	ReadFileErr bool     // 读取文件是否失败
}

// TooManyErr 当Parse太多语法错误的时候，终止
type TooManyErr struct {
	ErrNum int //  错误的数量
}

// CommentLine 当行的注释信息
type CommentLine struct {
	Str  string // 当行的注释内容
	Line int    // 第一个注释内容，出现的行号
	Col  int    // 第一个注释内容，出现的列号
}

// CommentInfo 单个注释片段信息，包含多行的注释信息
type CommentInfo struct {
	LineVec   []CommentLine // 多行的内容存储
	ShortFlag bool          // 是否是短注释，true表示短注释
	HeadFlag  bool          // 是否为头部注释， 例如一行中 --这样开头的就为头部注释
}

// GetRangeLoc 获取两个位置的范围，为[]
func GetRangeLoc(beginLoc, endLoc *Location) Location {
	return Location{
		StartLine:   beginLoc.StartLine,
		StartColumn: beginLoc.StartColumn,
		EndLine:     endLoc.EndLine,
		EndColumn:   endLoc.EndColumn,
	}
}

// GetRangeLocExcludeEnd 获取两个位置的范围，第二个位置不算,为[)
func GetRangeLocExcludeEnd(beginLoc, endLoc *Location) Location {
	return Location{
		StartLine:   beginLoc.StartLine,
		StartColumn: beginLoc.StartColumn,
		EndLine:     endLoc.StartLine,
		EndColumn:   endLoc.StartColumn,
	}
}

// CompareTwoLoc 对比两个loc是否一样的
func CompareTwoLoc(oneLoc, twoLoc *Location) bool {
	if oneLoc.StartLine != twoLoc.StartLine {
		return false
	}

	if oneLoc.EndLine != twoLoc.EndLine {
		return false
	}

	if oneLoc.StartColumn != twoLoc.StartColumn {
		return false
	}

	if oneLoc.EndColumn != twoLoc.EndColumn {
		return false
	}

	return true
}

// IsInitialLoc 判断位置信息是否为初始的
func (loc *Location) IsInitialLoc() bool {
	if loc.StartLine == 0 && loc.StartColumn == 0 && loc.EndLine == 0 && loc.EndColumn == 0 {
		return true
	}

	return false
}

// IsInLocStruct 判断传入的位置，是否在范围之内
func (loc *Location) IsInLocStruct(posLine int, posCh int) bool {
	if posLine < loc.StartLine {
		return false
	}

	if posLine > loc.EndLine {
		return false
	}

	if posLine == loc.StartLine && posCh < loc.StartColumn {
		return false
	}

	if posLine == loc.EndLine && posCh > loc.EndColumn {
		return false
	}

	return true
}

// IsContainLoc 判断当前 loc是否 包含传入的loc
func (loc Location) IsContainLoc(locOne Location) bool {
	if loc.StartLine > locOne.StartLine || loc.EndLine < locOne.EndLine {
		return false
	}

	if loc.StartLine == locOne.StartLine && loc.StartColumn > locOne.StartColumn {
		return false
	}

	if loc.EndLine == locOne.EndLine && loc.EndColumn < locOne.EndColumn {
		return false
	}

	return true
}

// IsBeforeLoc 判断当前 loc是否 在传入的loc之前
func (loc Location) IsBeforeLoc(locOne Location) bool {
	if loc.StartLine < locOne.StartLine {
		return true
	}

	if loc.StartLine == locOne.StartLine && loc.StartColumn <= locOne.StartColumn {
		return true
	}

	return false
}
