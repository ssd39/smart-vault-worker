package utils

func Prepend(dataToPrepend []byte, originalData []byte) []byte {
	// Create a new byte slice with the length of the original data plus the length of the data to prepend
	newData := make([]byte, len(dataToPrepend)+len(originalData))

	// Copy the data to prepend to the beginning of the new slice
	copy(newData, dataToPrepend)

	// Copy the original data after the prepended data
	copy(newData[len(dataToPrepend):], originalData)

	return newData
}
