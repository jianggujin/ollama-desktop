package util

// 如果有错误则触发panic
func Must(value interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return value
}
