package util

type Part []string

func LongestCommonPrefix(parts []Part) Part {
	var result Part

	maxIndex := 0
	for i, part := range parts {
		strLen := len(part)
		if i == 0 {
			maxIndex = strLen - 1
			result = part
			continue
		}

		if strLen-1 < maxIndex {
			maxIndex = strLen - 1
			result = result[:strLen]
		}

		for j := 0; j <= maxIndex && j < strLen; j++ {
			if part[j] != result[j] {
				maxIndex = j - 1
				result = part[:j]
			}
		}
	}

	return result
}
