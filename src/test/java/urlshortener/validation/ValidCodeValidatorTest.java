package urlshortener.validation;

import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.NullAndEmptySource;
import org.junit.jupiter.params.provider.ValueSource;

import static org.assertj.core.api.Assertions.assertThat;

class ValidCodeValidatorTest {

    private final ValidCodeValidator validator = new ValidCodeValidator();

    @Nested
    class ValidCodes {

        @ParameterizedTest
        @ValueSource(strings = {
                "0",
                "a",
                "code",
                "2137X",
                "zfGj3",
        })
        void testAcceptValidCodes(String code) {
            assertThat(validator.isValid(code, null)).isTrue();
        }
    }

    @Nested
    class InvalidCodes {

        @ParameterizedTest
        @NullAndEmptySource
        @ValueSource(
                strings = {
                        " ",
                        "\t",
                        "\n",
                        "  "})
        void testRejectEdgeCaseCodes(String code) {
            assertThat(validator.isValid(code, null)).isFalse();
        }


        @Test
        void testRejectTooLongCodes() {
            String longCode = "A".repeat(10);
            assertThat(validator.isValid(longCode, null)).isFalse();
        }

        @ParameterizedTest
        @ValueSource(
                strings = {
                        "code!",
                        "$&(@#",
                        "_"})
        void testRejectNonAlphanumericCodes(String code) {
            assertThat(validator.isValid(code, null)).isFalse();
        }
    }
}
