package urlshortener.url;

import jakarta.persistence.*;

import java.time.Instant;

@Entity
@Table(name = "shortened_urls")
public class ShortenedUrl {
    private final Instant createdAt;
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    private String originalUrl;
    private long clickCounter = 0;
    private Instant lastClickedAt;

    public ShortenedUrl() {
        this.createdAt = Instant.now();
        this.lastClickedAt = null;
    }

    public Long getId() {
        return id;
    }

    public String getOriginalUrl() {
        return originalUrl;
    }

    public void setOriginalUrl(String originalUrl) {
        this.originalUrl = originalUrl;
    }

    public long getClickCounter() {
        return clickCounter;
    }

    public void incrementClickCounter() {
        clickCounter++;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public Instant getLastClickedAt() {
        return lastClickedAt;
    }

    public void setLastClickedAt(Instant lastClickedAt) {
        this.lastClickedAt = lastClickedAt;
    }
}
