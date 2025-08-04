package urlshortener.transcoding;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class Base62ConverterTest {

    @Test
    void testEncodeDecode() {
        long originalId = 21;

        String encoded = Base62Converter.encode(originalId);
        long decoded = Base62Converter.decode(encoded);

        assertThat(decoded).isEqualTo(originalId);
    }

    @ParameterizedTest
    @CsvSource({"1, 1",
            "62, 10",
            "123, 1z"})
    void testEncoding(long id, String expectedBase62) {
        String result = Base62Converter.encode(id);

        assertThat(result).isEqualTo(expectedBase62);
    }

    @Test
    void testDecoding() {
        String base62 = "1";

        long result = Base62Converter.decode(base62);

        assertThat(result).isEqualTo(1);
    }

    @Test
    void shouldThrowExceptionForInvalidCharacter() {
        assertThatThrownBy(() -> Base62Converter.decode("abc$"))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("Invalid character in Base62 string: $");
    }
}