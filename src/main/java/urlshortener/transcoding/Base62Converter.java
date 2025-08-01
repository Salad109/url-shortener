package urlshortener.transcoding;

public class Base62Converter {

    private static final String BASE62_ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    private static final int BASE = BASE62_ALPHABET.length();

    public static String encode(long id) {
        StringBuilder encoded = new StringBuilder();
        do {
            encoded.append(BASE62_ALPHABET.charAt((int) (id % BASE)));
            id /= BASE;
        } while (id > 0);
        return encoded.reverse().toString();
    }

    public static long decode(String base62) {
        long decoded = 0;
        for (char c : base62.toCharArray()) {
            int index = BASE62_ALPHABET.indexOf(c);
            if (index == -1) {
                throw new IllegalArgumentException("Invalid character in Base62 string: " + c);
            }
            decoded = decoded * BASE + index;
        }
        return decoded;
    }
}
