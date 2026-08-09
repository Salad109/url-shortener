package transcoding

// Encode takes a sequentially-generated ID and obfuscates it into a base62 string
func Encode(id int64) string {
	scrambled := Scramble(id)
	encoded := EncodeBase62(scrambled)
	return encoded
}

// Decode takes the obfuscated short code string and decodes it into the sequential ID
func Decode(encoded string) (int64, error) {
	decoded, err := DecodeBase62(encoded)
	if err != nil {
		return 0, err
	}
	unscrambled := Unscramble(decoded)
	return unscrambled, nil
}
