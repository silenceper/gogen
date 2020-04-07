package orm

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	sh "github.com/codeskyblue/go-sh"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ngaut/log"
	"github.com/silenceper/gogen/pkg"
	"github.com/silenceper/gogen/pkg/funcs"
	"github.com/silenceper/gogen/util"
	"gopkg.in/yaml.v1"
)

var dbConfigs []*dbConfig

type dbConfig struct {
	Pkg        string       `yaml:"pkg"`
	Database   string       `yaml:"database"`
	DataSource string       `yaml:"dataSource"`
	Table      []*tableInfo `yaml:"table"`
	db         *sql.DB
}

type tableInfo struct {
	Name   string `yaml:"name"`
	Prefix string `yaml:"prefix"`
}

var typeMap = [][]string{
	{"tinyint", "byte", "db.NullInt64"},
	{"smallint", "int32", "db.NullInt64"},
	{"int", "int32", "db.NullInt64"},
	{"bigint", "int64", "db.NullInt64"},
	{"varchar", "string", "db.NullString"},
	{"char", "string", "db.NullString"},
	{"text", "string", "db.NullString"},
	{"tinytext", "string", "db.NullString"},
	{"datetime", "time.Time", "db.NullTime"},
	{"date", "time.Time", "db.NullTime"},
	{"timestamp", "time.Time", "db.NullTime"},
	{"time", "time.Time", "db.NullTime"},
	{"decimal", "float64", "db.NullFloat64"},
	{"enum", "string", "db.NullString"},
	{"bit", "db.Bit", "db.NullBit"},
}

// ColumnInfo table column info
type ColumnInfo struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default *string
	Extra   string

	AutoIncrement bool
	Unique        bool
	GoType        string
}

// GenOrm 生成orm文件
func GenOrm(ormFile string) {
	ormFile, err := filepath.Abs(ormFile)
	if err != nil {
		panic(err)
	}
	if !util.CheckFileExist(ormFile) {
		panic("配置文件不存在：" + ormFile)
	}
	fileBytes, err := ioutil.ReadFile(ormFile)
	if err != nil {
		panic(err)
	}

	if err = yaml.Unmarshal(fileBytes, &dbConfigs); err != nil {
		panic(err)
	}
	for _, config := range dbConfigs {
		//获取数据库连接
		db, err := sql.Open("mysql", config.DataSource)
		config.db = db
		if err != nil {
			panic(err)
		}
		defer db.Close()
		//生成目录
		targetPath := fmt.Sprintf("./pkg/model/%s", config.Pkg)
		err = util.MkDirIfNotExist(targetPath)
		if err != nil {
			panic(err)
		}
		//生成gen_db.go
		dbTemplateRender(config, targetPath)

		//生成表操作
		for _, table := range config.Table {
			tableTemplateRender(config, table, targetPath)
		}
		sh.Command("gofmt", "-w", ".", sh.Dir(targetPath)).Run()
	}
}

func dbTemplateRender(config *dbConfig, targetPath string) {
	tableTplFile := fmt.Sprintf("golang/pkg/model/database/db.go.tpl")
	tplBytes, err := pkg.Asset(tableTplFile)
	if err != nil {
		panic(err)
	}
	tpl, err := template.New(config.Database).Funcs(funcs.FuncMap).Parse(string(tplBytes))
	if err != nil {
		panic(err)
	}
	metaInfo := map[string]interface{}{
		"Database": config.Database,
		"Pkg":      config.Pkg,
	}
	filePath := fmt.Sprintf("%s/gen_db.go", targetPath)
	renderFile(filePath, tpl, metaInfo)
}

func tableTemplateRender(config *dbConfig, tableInfo *tableInfo, targetPath string) {
	tableName := tableInfo.Name
	metaInfo := getTableMetaInfo(config.db, tableInfo)

	tableTplFile := fmt.Sprintf("golang/pkg/model/database/table.go.tpl")
	tplBytes, err := pkg.Asset(tableTplFile)
	if err != nil {
		panic(err)
	}
	funcs.FuncMap["getTableFieldNames"] = GetTableFieldNames
	tpl, err := template.New(tableName).Funcs(funcs.FuncMap).Parse(string(tplBytes))
	if err != nil {
		panic(err)
	}
	metaInfo["DBName"] = config.Database
	metaInfo["Pkg"] = config.Pkg
	filePath := fmt.Sprintf("%s/gen_%s.go", targetPath, tableName)
	renderFile(filePath, tpl, metaInfo)
}

func renderFile(filePath string, tpl *template.Template, metaInfo map[string]interface{}) {
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

//获取表字段元信息
func getTableMetaInfo(db *sql.DB, tableInfo *tableInfo) map[string]interface{} {
	tableName := tableInfo.Name
	row, err := db.Query("show columns from " + tableName)
	if err != nil {
		panic(err)
	}

	var (
		primaryKey              = ""
		primaryKeyType          = ""
		primaryKeyExtra         = ""
		primaryKeyAutoIncrement = false
		columnInfoList          []*ColumnInfo
	)
	//是否已经有主键了
	var hasPRI bool
	var hasMultiPRI bool
	//var columns []*ColumnInfo
	for row.Next() {
		//var column ColumnInfo
		c := new(ColumnInfo)
		if err = row.Scan(&c.Field, &c.Type, &c.Null, &c.Key, &c.Default, &c.Extra); err != nil {
			panic(err)
		}
		c.GoType = toGoType(c.Type, c.Null)

		if strings.Contains(c.Extra, "auto_increment") {
			c.AutoIncrement = true
		}

		if c.Key == "UNI" {
			c.Unique = true
		}

		columnInfoList = append(columnInfoList, c)

		//TODO: 暂时只处理只有一个主键字段的情况
		if c.Key == "PRI" {
			if hasPRI {
				hasMultiPRI = true
			}
			hasPRI = true
			primaryKey = c.Field
			primaryKeyType = c.GoType
			primaryKeyExtra = c.Extra
			if c.Extra == "auto_increment" {
				primaryKeyAutoIncrement = true
			}
		}
	}
	if hasMultiPRI {
		primaryKey = ""
		primaryKeyType = ""
		primaryKeyExtra = ""
		primaryKeyAutoIncrement = false
	}
	prettyTableName := strings.TrimPrefix(tableName, tableInfo.Prefix)

	fieldStringList := ""
	for _, column := range columnInfoList {
		fieldStringList = fmt.Sprintf("%s,`%s`", fieldStringList, column.Field)
	}
	fieldStringList = strings.TrimLeft(fieldStringList, ",")

	return map[string]interface{}{
		"TableName":               tableName,
		"TablePrefix":             tableInfo.Prefix,
		"PrettyTableName":         prettyTableName,
		"PrimaryKey":              primaryKey,
		"PrimaryKeyType":          primaryKeyType,
		"PrimaryKeyExtra":         primaryKeyExtra,
		"PrimaryKeyAutoIncrement": primaryKeyAutoIncrement,
		"Columns":                 columnInfoList,
		"FieldStringList":         fieldStringList,
	}
}

func toGoType(s, null string) string {
	for _, v := range typeMap {
		if strings.HasPrefix(s, v[0]) {
			if null == "YES" {
				return v[2]
			}
			return v[1]
		}
	}
	panic("unsupport type " + s)
}

//GetTableFieldNames 根据ColumnInfo 数组返回根据逗号拼接的字段
func GetTableFieldNames(args []*ColumnInfo) string {
	names := []string{}
	for _, a := range args {
		names = append(names, fmt.Sprintf("`%s`", a.Field))
	}
	return strings.Join(names, ", ")
}
