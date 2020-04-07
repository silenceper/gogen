package orm

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/silenceper/log"
)

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

//GormLogger gorm logger
type GormLogger struct {
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func getLastFile(s string) string {
	arr := strings.Split(s, "/")
	return arr[len(arr)-1]
}

//Print 输出sql执行信息
func (logger *GormLogger) Print(values ...interface{}) {
	//pretty.Println(values)
	if len(values) <= 1 {
		return
	}

	var (
		sql             string
		formattedValues []string
		level           = values[0]
		source          = fmt.Sprintf("%v", values[1])
		duration        string
	)
	if level == "sql" {
		// duration
		duration = fmt.Sprintf("%.2fms", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0)

		// sql
		for _, value := range values[4].([]interface{}) {
			indirectValue := reflect.Indirect(reflect.ValueOf(value))
			if indirectValue.IsValid() {
				value = indirectValue.Interface()
				if t, ok := value.(time.Time); ok {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
				} else if b, ok := value.([]byte); ok {
					if str := string(b); isPrintable(str) {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
					} else {
						formattedValues = append(formattedValues, "'<binary>'")
					}
				} else if r, ok := value.(driver.Valuer); ok {
					if value, err := r.Value(); err == nil && value != nil {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					} else {
						formattedValues = append(formattedValues, "NULL")
					}
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			} else {
				formattedValues = append(formattedValues, "NULL")
			}
		}

		// differentiate between $n placeholders or else treat like ?
		if numericPlaceHolderRegexp.MatchString(values[3].(string)) {
			sql = values[3].(string)
			for index, value := range formattedValues {
				placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
				sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
			}
		} else {
			formattedValuesLength := len(formattedValues)
			for index, value := range sqlRegexp.Split(values[3].(string), -1) {
				sql += value
				if index < formattedValuesLength {
					sql += formattedValues[index]
				}
			}
		}
		rowsAffected := strconv.FormatInt(values[5].(int64), 10)
		log.Debugf("sql:%v, duration:%v, rows affected:%v, source:%v", sql, duration, rowsAffected, getLastFile(source))
	} else {
		log.Errorf("%v", values[2:])
	}
}
