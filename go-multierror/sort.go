package multierror

// 实现 Errors 的长度
func (e *Error) Len() int {
	if e == nil {
		return 0
	}

	return len(e.Errors)
}

// 交换两个元素的位置
func (e *Error) Swap(i, j int) {
	e.Errors[i], e.Errors[j] = e.Errors[j], e.Errors[i]
}

// 比较两个元素的大小
func (e *Error) Less(i, j int) bool {
	return e.Errors[i].Error() < e.Errors[j].Error()
}
