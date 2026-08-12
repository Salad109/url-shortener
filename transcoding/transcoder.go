package transcoding

import "errors"

// MaxCodeLength is the longest code Encode can produce, since MaxValue is 62^5 - 1.
const MaxCodeLength = 5

// Encode takes a sequentially-generated ID and obfuscates it into a base62 string
func Encode(id int64) string {
	scrambled := Scramble(id)
	encoded := EncodeBase62(scrambled)
	return encoded
}

// Decode takes the obfuscated short code string and decodes it into the sequential ID
func Decode(encoded string) (int64, error) {
	if len(encoded) > MaxCodeLength {
		return 0, errors.New("short code is too long")
	}
	if encoded == "" {
		return 0, errors.New("short code is empty")
	}
	if encoded[0] == '0' {
		return 0, errors.New("short code has a leading zero")
	}

	decoded, err := DecodeBase62(encoded)
	if err != nil {
		return 0, err
	}
	if decoded >= MaxValue {
		return 0, errors.New("short code is out of range")
	}

	unscrambled := Unscramble(decoded)
	return unscrambled, nil
}
