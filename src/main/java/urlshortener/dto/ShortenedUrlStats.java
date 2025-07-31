package urlshortener.dto;

public record ShortenedUrlStats(String code, String originalUrl, long clickCounter) {
}
