package raml

// append `str` to `arr` if `str` not exist in `arr`
func appendStrNotExist(str string, arr []string) []string {
	// check if a `str` exist in `arr`
	isStrInArr := func(str string, arr []string) bool {
		for _, s := range arr {
			if str == s {
				return true
			}
		}
		return false
	}

	if !isStrInArr(str, arr) {
		arr = append(arr, str)
	}
	return arr
}
