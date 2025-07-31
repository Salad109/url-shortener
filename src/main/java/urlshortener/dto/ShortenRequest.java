package urlshortener.dto;

import jakarta.validation.constraints.NotBlank;
import urlshortener.validation.ValidUrl;

public record ShortenRequest(@NotBlank @ValidUrl String originalUrl) {
}
