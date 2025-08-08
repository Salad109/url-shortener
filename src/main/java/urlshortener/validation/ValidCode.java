package urlshortener.validation;

import jakarta.validation.Constraint;
import jakarta.validation.Payload;

import java.lang.annotation.*;

@Documented
@Constraint(validatedBy = ValidCodeValidator.class)
@Target({ElementType.FIELD, ElementType.PARAMETER})
@Retention(RetentionPolicy.RUNTIME)
public @interface ValidCode {
    String message() default "Code must be alphanumeric and up to 5 characters long";

    Class<?>[] groups() default {};

    Class<? extends Payload>[] payload() default {};
}
