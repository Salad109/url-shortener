package urlshortener.validation;

import jakarta.validation.ConstraintValidator;
import jakarta.validation.ConstraintValidatorContext;

import java.util.regex.Pattern;

public class ValidCodeValidator implements ConstraintValidator<ValidCode, String> {

    private static final int MAX_CODE_LENGTH = 5;
    private static final Pattern PATTERN = Pattern.compile("^[a-zA-Z0-9]+$");

    @Override
    public boolean isValid(String code, ConstraintValidatorContext context) {
        if (code == null || code.isBlank()) {
            return false;
        }

        if (code.length() > MAX_CODE_LENGTH) {
            return false;
        }

        return PATTERN.matcher(code).matches();
    }
}
