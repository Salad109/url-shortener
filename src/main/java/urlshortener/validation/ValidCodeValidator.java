package urlshortener.validation;

import jakarta.validation.ConstraintValidator;
import jakarta.validation.ConstraintValidatorContext;

public class ValidCodeValidator implements ConstraintValidator<ValidCode, String> {

    private static final int MAX_CODE_LENGTH = 5;

    @Override
    public boolean isValid(String code, ConstraintValidatorContext context) {
        if (code == null || code.isBlank()) {
            return false;
        }

        if (code.length() > MAX_CODE_LENGTH) {
            return false;
        }

        // Alphanumeric check
        for (char c : code.toCharArray()) {
            if (!((c >= '0' && c <= '9') ||
                    (c >= 'A' && c <= 'Z') ||
                    (c >= 'a' && c <= 'z'))) {
                return false;
            }
        }
        return true;
    }
}
