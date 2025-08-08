package urlshortener.url;

import java.time.Instant;

public record ShortenedUrlData(
        String originalUrl,
        long clickCounter,
        Instant createdAt,
        Instant lastClickedAt
) {
    public ShortenedUrlData(String originalUrl) {
        this(originalUrl, 0, Instant.now(), null);
    }

    public ShortenedUrlData incrementClicks() {
        return new ShortenedUrlData(originalUrl, clickCounter + 1, createdAt, Instant.now());
    }
}