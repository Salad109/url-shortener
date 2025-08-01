package urlshortener.cleanup;

import jakarta.transaction.Transactional;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import urlshortener.url.ShortenedUrlRepository;

import java.time.Instant;
import java.time.temporal.ChronoUnit;

@Service
public class ExpiredUrlCleanupService {

    private static final Logger log = LoggerFactory.getLogger(ExpiredUrlCleanupService.class);
    private final ShortenedUrlRepository repository;

    public ExpiredUrlCleanupService(ShortenedUrlRepository repository) {
        this.repository = repository;
    }

    @Scheduled(fixedRate = 15 * 60 * 1000) // 15 minutes
    @Transactional
    public void cleanupExpiredUrls() {
        log.info("Starting cleanup of expired URLs");

        Instant fiveMinutesAgo = Instant.now().minus(15, ChronoUnit.MINUTES);

        int deletedCount = repository.deleteExpiredUrls(fiveMinutesAgo);

        log.info("Cleanup completed. Deleted {} expired URLs", deletedCount);
    }
}