package urlshortener.transcoding;

public class Base62Converter {

    private static final String BASE62_ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    private static final int BASE = BASE62_ALPHABET.length();

    public static String encode(long id) {
        if (id < 0) {
            throw new IllegalArgumentException("ID must be non-negative");
        }
        StringBuilder encoded = new StringBuilder();
        do {
            encoded.append(BASE62_ALPHABET.charAt((int) (id % BASE)));
            id /= BASE;
        } while (id > 0);
        return encoded.reverse().toString();
    }

    public static long decode(String base62) {
        if (base62 == null || base62.isEmpty()) {
            throw new IllegalArgumentException("Base62 string must not be null or empty");
        }
        long decoded = 0;
        for (char c : base62.toCharArray()) {
            int index = BASE62_ALPHABET.indexOf(c);
            if (index < 0) {
                throw new IllegalArgumentException("Invalid character in Base62 string: " + c);
            }
            decoded = decoded * BASE + index;
        }
        return decoded;
    }
}
