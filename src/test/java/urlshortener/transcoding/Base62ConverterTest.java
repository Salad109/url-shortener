package urlshortener.transcoding;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class Base62ConverterTest {

    @ParameterizedTest
    @CsvSource({"1, 1",
            "62, 10",
            "123, 1z"})
    void testEncodeDecode(long originalId, String expectedBase62) {
        String encoded = Base62Converter.encode(originalId);

        assertThat(encoded).isEqualTo(expectedBase62);

        long decoded = Base62Converter.decode(encoded);

        assertThat(decoded).isEqualTo(originalId);
    }

    @Test
    void testInvalidCharacterException() {
        assertThatThrownBy(() -> Base62Converter.decode("abc$"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("Invalid character in Base62 string: $");
    }
}