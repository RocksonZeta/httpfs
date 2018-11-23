package stringutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"unicode/utf8"

	// "github.com/bradfitz/slice"

	// slugify "github.com/mozillazg/go-slugify"

	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func JoinInts(is []int) string {
	s := ""
	for _, iv := range is {
		s += strconv.Itoa(iv)
		s += ","
	}
	if "" == s {
		return "null"
	}
	return s[0 : len(s)-1]
}

func JoinInt64s(is []int64) string {
	s := ""
	for _, iv := range is {
		s += strconv.FormatInt(iv, 10)
		s += ","
	}
	if "" == s {
		return "null"
	}
	return s[0 : len(s)-1]
}
func SliceStr2Int(ss []string) []int {
	if len(ss) <= 0 {
		return nil
	}
	r := make([]int, 0, len(ss))
	for _, s := range ss {
		if "" == s {
			continue
		}
		v, _ := strconv.Atoi(s)
		r = append(r, v)
	}
	return r
}

func JoinStrs(ks []string, sep string) string {
	r := ""
	for i, v := range ks {
		r += v
		if i < len(ks)-1 {
			r += sep
		}
	}
	return r
}

func ToInts(ss []string) []int {
	r := make([]int, len(ss))
	for i, v := range ss {
		r[i], _ = strconv.Atoi(v)
	}
	return r
}

func ToInt(s string, dv ...int) int {
	v, err := strconv.Atoi(s)
	if nil != err {
		if len(dv) > 0 {
			return dv[0]
		}
		return 0
	}
	return v
}

func IntsSubstract(a, b []int) []int {
	if nil == a {
		return nil
	}
	if nil == b {
		return a
	}
	var r []int
	m := make(map[int]bool)
	for _, v := range b {
		m[v] = true
	}
	for _, v := range a {
		if !m[v] {
			r = append(r, v)
		}
	}
	return r
}

func IntsIntersect(a, b []int) []int {
	if nil == a {
		return nil
	}
	if nil == b {
		return nil
	}
	var r []int
	m := make(map[int]bool)
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		if m[v] {
			r = append(r, v)
		}
	}
	return r
}

func ToString(value interface{}) (string, error) {
	switch value.(type) {
	case error:
		e, _ := value.(error)
		if e != nil {
			panic(e)
		}
	case bool:
		return strconv.FormatBool(value.(bool)), nil
	case string:
		return value.(string), nil
	case int:
		return strconv.Itoa(value.(int)), nil
	case int64:
		return strconv.FormatInt(value.(int64), 10), nil
	case float32:
		return fmt.Sprintf("%f", value), nil
	case float64:
		return fmt.Sprintf("%f", value), nil
	}
	return "", errors.New("不支持此类型转换")
}

func IsUtf8(s []byte) bool {
	return utf8.Valid(s)
}
func GbkToUtf8(s []byte) ([]byte, error) {
	if utf8.Valid(s) {
		return s, nil
	}
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GB18030.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return s, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GB18030.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return s, e
	}
	return d, nil
}

func compare(a int, as []int, less bool, cb func(a, b int) bool) []int {
	var r []int
	for _, v := range as {
		if less {
			if cb(a, v) {
				r = append(r, v)
			}
		} else {
			if !cb(a, v) {
				r = append(r, v)
			}
		}

	}
	return r
}

func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func JoinSqlFields(prefix string, fields ...string) string {
	if len(fields) <= 0 {
		return ""
	}
	r := ""
	for _, v := range fields {
		if prefix != "" {

			r += prefix + "." + v + ","
		} else {
			r += v + ","
		}
	}

	return r[0 : len(r)-1]
}

// func SortStrs(strs []string) {
// 	slice.Sort(strs, func(i, j int) bool {
// 		return strings.Compare(strs[i], strs[j]) < 0
// 	})
// }

func IndexStrs(strs []string, str string) int {
	if len(strs) <= 0 {
		return -1
	}
	for i, s := range strs {
		if s == str {
			return i
		}
	}
	return -1
}

func DeHtmlTag(html string) string {
	//<a> </a>
	re := regexp.MustCompile("</?\\w.*?>")
	return re.ReplaceAllString(html, "")
}
func JsonPretty(v interface{}) string {
	bs, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Println(err)
	}
	return string(bs)
}
func Json(v interface{}) string {
	bs, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
	}
	return string(bs)
}

func In(n int, ints []int) bool {
	for _, v := range ints {
		if n == v {
			return true
		}
	}
	return false
}

//FileNameAppend hello.jpg,_1 -> hello_1.jpg
func FileNameAppend(filename, subname string) string {
	i := strings.LastIndex(filename, ".")
	if i == -1 {
		return filename + subname
	}
	return filename[0:i] + subname + filename[i:]
}
