package indicators

func safeWindow(window int) int {
	if window <= 0 {
		return 1
	}
	return window
}
