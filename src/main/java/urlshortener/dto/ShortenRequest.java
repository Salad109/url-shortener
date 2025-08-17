package urlshortener.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import urlshortener.validation.ValidUrl;

public record ShortenRequest(
        @Schema(
                description = "The original URL to be shortened",
                example = "https://example.com/very-long-url-very-ugly",
                requiredMode = Schema.RequiredMode.REQUIRED)
        @NotBlank
        @ValidUrl
        String originalUrl) {
}
