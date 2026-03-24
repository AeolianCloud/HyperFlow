package timeutil

import "time"

var shanghaiLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

// ShanghaiLocation 返回应用内统一使用的固定上海时区。
func ShanghaiLocation() *time.Location {
	return shanghaiLocation
}

// NowShanghai 返回固定上海时区下的当前时间。
func NowShanghai() time.Time {
	return time.Now().In(shanghaiLocation)
}

// InShanghai 将任意时间统一转换为固定上海时区表示。
func InShanghai(t time.Time) time.Time {
	return t.In(shanghaiLocation)
}
