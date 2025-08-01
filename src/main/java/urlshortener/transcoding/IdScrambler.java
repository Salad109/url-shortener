package urlshortener.transcoding;

public class IdScrambler {
    private static final long MAX_VALUE = 916_132_831L; // 65^5 - 1
    private static final long LARGE_PRIME = 687_194_767L;
    private static final long INVERSE = 490_572_491L; // Pre-calculated using Extended Euclidean Algorithm

    public static long encode(long id) {
        if (id <= 0 || id >= MAX_VALUE) {
            throw new IllegalArgumentException("ID must be between 1 and " + (MAX_VALUE - 1));
        }
        return (id * LARGE_PRIME) % MAX_VALUE;
    }

    public static long decode(long scrambledId) {
        if (scrambledId <= 0 || scrambledId >= MAX_VALUE) {
            throw new IllegalArgumentException("Scrambled ID out of valid range");
        }
        return (scrambledId * INVERSE) % MAX_VALUE;
    }
}