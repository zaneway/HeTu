package util

import "time"

const DateTime = "2006-01-02 15:04:05"

// yyyy-MM-dd HH:mm:ss
const FormatStr = "060102150405Z0700"

// toBeijingTime 转换为北京时间
func ToBeijingTime(t time.Time) time.Time {
	loc := time.FixedZone("CST", 8*3600)
	return t.In(loc)
}
