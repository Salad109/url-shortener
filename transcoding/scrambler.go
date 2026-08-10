package transcoding

const MaxValue = 916_132_831
const LargePrime = 687_194_767
const Inverse = 490_572_491

func Scramble(id int64) int64 {
	return (id * LargePrime) % MaxValue
}

func Unscramble(id int64) int64 {
	return (id * Inverse) % MaxValue
}
