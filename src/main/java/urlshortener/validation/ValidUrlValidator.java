package urlshortener.validation;

import jakarta.validation.ConstraintValidator;
import jakarta.validation.ConstraintValidatorContext;

public class ValidUrlValidator implements ConstraintValidator<ValidUrl, String> {

    private static final int MAX_URL_LENGTH = 2048;

    @Override
    public boolean isValid(String url, ConstraintValidatorContext context) {
        if (url == null || url.isBlank()) {
            return false;
        }

        if (url.length() > MAX_URL_LENGTH) {
            return false;
        }

        return url.startsWith("http://") || url.startsWith("https://");
    }
}