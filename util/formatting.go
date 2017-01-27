package util

func PadRight(str string, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}

func PadLeft(str string, pad string, length int) string {
	for {
		str = pad + str
		if len(str) > length-1 {
			return str[0:length]
		}
	}
}
