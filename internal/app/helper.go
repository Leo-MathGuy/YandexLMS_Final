package app

func FloatPtr(n float64) *float64 {
	v := n
	return &v
}

func IntPtr(n int) *int {
	v := n
	return &v
}
