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

        return code.matches("^[a-zA-Z0-9]+$");
    }
}
