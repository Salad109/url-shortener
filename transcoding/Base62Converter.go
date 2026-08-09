package transcoding

import "fmt"

const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const Base = int64(len(Alphabet))

func EncodeBase62(id int64) string {
	encoded := ""
	for id > 0 {
		remainder := id % Base
		encoded = string(Alphabet[remainder]) + encoded
		id /= Base
	}

	return encoded
}

func DecodeBase62(encoded string) (int64, error) {
	decoded := int64(0)
	for _, char := range encoded {
		index := int64(-1)
		for j, alphabetChar := range Alphabet {
			if char == alphabetChar {
				index = int64(j)
				break
			}
		}
		if index == -1 {
			return 0, fmt.Errorf("invalid character in Base62 string")
		}
		decoded = decoded*Base + index
	}

	return decoded, nil
}
