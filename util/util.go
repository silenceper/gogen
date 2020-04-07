package util

import (
	"os"
	"text/template"

	"github.com/silenceper/log"
)

//CheckFileExist 检查文件是否存在
func CheckFileExist(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

//MkDirsIfNotExist 批量创建文件夹
func MkDirsIfNotExist(dirs []string) {
	for _, dir := range dirs {
		err := MkDirIfNotExist(dir)
		if err != nil {
			panic("mkdir " + dir + " error")
		}
	}
}

// MkDirIfNotExist 如果指定文件夹不存在则创建
func MkDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

//RenderFile 生成文件
func RenderFile(filePath string, tpl *template.Template, metaInfo map[string]interface{}) {
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = tpl.Execute(f, metaInfo)
	if err != nil {
		panic(err)
	}
	log.Infof("[SUCCESS] 文件生成成功： %s ", filePath)
}
