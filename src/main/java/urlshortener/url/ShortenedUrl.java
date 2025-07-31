package urlshortener.url;

import jakarta.persistence.*;

@Entity
@Table(name = "shortened_urls")
public class ShortenedUrl {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    private String originalUrl;

    private int clickCounter = 0;

    public ShortenedUrl() {
    }

    public ShortenedUrl(String originalUrl) {
        this.originalUrl = originalUrl;
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

    public int getClickCounter() {
        return clickCounter;
    }

    public void incrementClickCounter() {
        clickCounter++;
    }
}
