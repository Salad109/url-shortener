package urlshortener.dto;

import io.swagger.v3.oas.annotations.media.Schema;

import java.time.Instant;

public record ShortenedUrlStats(
        @Schema(
                description = "Short code associated with the original URL",
                example = "kVOkZ")
        String code,
        @Schema(
                description = "Original URL that was shortened",
                example = "https://example.com/very-long-url-very-ugly")
        String originalUrl,
        @Schema(
                description = "Total number of clicks on the shortened URL",
                example = "21")
        long clickCounter,
        @Schema(
                description = "Timestamp when the shortened URL was created",
                example = "2025-08-17T12:34:56Z")
        Instant createdAt,
        @Schema(
                description = "Timestamp when the code was last used to redirect",
                example = "2025-08-17T12:34:56Z")
        Instant lastClickedAt) {
}
