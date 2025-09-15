package urlshortener.dto;

import io.swagger.v3.oas.annotations.media.Schema;

public record ShortenResponse(
        @Schema(
                description = "Generated short code related to the original URL",
                example = "kVOkZ")
        String shortCode) {
}
