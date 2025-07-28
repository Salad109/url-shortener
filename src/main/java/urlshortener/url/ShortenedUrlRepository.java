package urlshortener.url;

import org.springframework.data.jpa.repository.JpaRepository;

public interface ShortenedUrlRepository extends JpaRepository<ShortenedUrl, Long> {
}
