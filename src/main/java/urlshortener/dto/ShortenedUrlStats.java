package urlshortener.dto;

import java.time.Instant;

public record ShortenedUrlStats(
        String code,
        String originalUrl,
        long clickCounter,
        Instant createdAt,
        Instant lastClickedAt) {
}
