package urlshortener.url;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.Instant;

public interface ShortenedUrlRepository extends JpaRepository<ShortenedUrl, Long> {

    @Modifying
    @Query("DELETE FROM ShortenedUrl WHERE " +
            "(lastClickedAt IS NOT NULL AND lastClickedAt < :cutoff) OR " +
            "(lastClickedAt IS NULL AND createdAt < :cutoff)")
    int deleteExpiredUrls(@Param("cutoff") Instant cutoff);
}