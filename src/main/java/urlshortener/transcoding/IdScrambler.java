package urlshortener.transcoding;

public class IdScrambler {

    private static final long MAX_VALUE = 916_132_831L; // 62^5 - 1
    private static final long LARGE_PRIME = 687_194_767L;
    private static final long INVERSE = 490_572_491L; // Pre-calculated using Extended Euclidean Algorithm

    private IdScrambler() {
    }

    public static long encode(long id) {
        return (id * LARGE_PRIME) % MAX_VALUE;
    }

    public static long decode(long scrambledId) {
        return (scrambledId * INVERSE) % MAX_VALUE;
    }
}