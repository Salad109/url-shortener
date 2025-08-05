package urlshortener.transcoding;

import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.assertj.core.api.Assertions.assertThat;

class IdScramblerTest {

    @ParameterizedTest
    @CsvSource({
            "0, 0",
            "1, 687194767",
            "2, 458256703"
    })
    void testEncodeDecode(long originalId, long expectedScrambledId) {
        long encoded = IdScrambler.encode(originalId);
        assertThat(encoded).isEqualTo(expectedScrambledId);

        long decoded = IdScrambler.decode(encoded);
        assertThat(decoded).isEqualTo(originalId);
    }
}
