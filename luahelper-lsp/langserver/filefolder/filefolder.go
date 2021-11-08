package filefolder

import "os"

// IsDirExist 判断所给文件夹是否存在
func IsDirExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}

	return s.IsDir()
}

// IsFileExist 判断所给文件夹是否存在
func IsFileExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !s.IsDir()
}
