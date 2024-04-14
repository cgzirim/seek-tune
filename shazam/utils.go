package shazam

// deduplicate returns a list of unique integers from the given array.
// The order of the given array is not preserved in the result.
func deduplicate(array []int) []int {
	uniqueMap := make(map[int]struct{})

	for _, num := range array {
		uniqueMap[num] = struct{}{}
	}

	var uniqueList []int
	for num := range uniqueMap {
		uniqueList = append(uniqueList, num)
	}

	return uniqueList
}
