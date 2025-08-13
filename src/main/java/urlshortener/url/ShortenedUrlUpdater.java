package urlshortener.url;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;

import java.time.Duration;
import java.time.Instant;

@Component
public class ShortenedUrlUpdater {

    private static final Logger log = LoggerFactory.getLogger(ShortenedUrlUpdater.class);
    private final RedisTemplate<String, byte[]> redisTemplate;

    public ShortenedUrlUpdater(RedisTemplate<String, byte[]> redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    @Async
    public void updateUrlStats(String code, urlshortener.proto.ShortenedUrl.ShortenedUrlData data) {
        urlshortener.proto.ShortenedUrl.ShortenedUrlData updated = data.toBuilder()
                .setClickCounter(data.getClickCounter() + 1)
                .setLastClickedAt(Instant.now().toEpochMilli())
                .build();

        redisTemplate.opsForValue().set(code, updated.toByteArray(), Duration.ofMinutes(5));
        log.debug("Updated click stats for code: {}", code);
    }
}
