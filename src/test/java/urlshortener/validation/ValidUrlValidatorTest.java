package urlshortener.validation;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.NullAndEmptySource;
import org.junit.jupiter.params.provider.ValueSource;

import static org.assertj.core.api.Assertions.assertThat;

class ValidUrlValidatorTest {

    private final ValidUrlValidator validator = new ValidUrlValidator();

    @ParameterizedTest
    @ValueSource(strings = {
            "http://example.com",
            "https://example.com",
            "https://very-very-long-url.com/it-sure-is-very-long-and-ugly/1234567890/goober",
    })
    void testAcceptValidUrls(String validUrl) {
        assertThat(validator.isValid(validUrl, null)).isTrue();
    }

    @ParameterizedTest
    @NullAndEmptySource // null, ""
    @ValueSource(strings = {" ", "\t", "\n"})
    void testRejectEdgeCaseUrls(String invalidUrl) {
        assertThat(validator.isValid(invalidUrl, null)).isFalse();
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "example.com",
            "www.example.com",
            "C:\\homework"
    })
    void testRejectNonHttpUrls(String invalidUrl) {
        assertThat(validator.isValid(invalidUrl, null)).isFalse();
    }

    @Test
    void testRejectTooLongUrls() {
        String longUrl = "https://" + "A".repeat(2048);

        assertThat(validator.isValid(longUrl, null)).isFalse();
    }
}