package urlshortener.url;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.exception.UrlSerializationException;
import urlshortener.proto.ShortenedUrl;

import java.time.Duration;
import java.time.Instant;

@Service
public class ShortenedUrlService {

    private static final Logger log = LoggerFactory.getLogger(ShortenedUrlService.class);
    private final RedisTemplate<String, byte[]> redisTemplate;
    private final IdGenerator idGenerator;
    private final ShortenedUrlUpdater shortenedUrlUpdater;

    public ShortenedUrlService(RedisTemplate<String, byte[]> redisTemplate, IdGenerator idGenerator, ShortenedUrlUpdater shortenedUrlUpdater) {
        this.redisTemplate = redisTemplate;
        this.idGenerator = idGenerator;
        this.shortenedUrlUpdater = shortenedUrlUpdater;
    }

    public String shortenUrl(String originalUrl) {
        String code = idGenerator.generateCode();

        ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.newBuilder()
                .setOriginalUrl(originalUrl)
                .setClickCounter(0)
                .setCreatedAt(Instant.now().toEpochMilli())
                .setLastClickedAt(0)
                .build();

        log.info("Shortening URL: {} with code: {}", originalUrl, code);
        redisTemplate.opsForValue().set(code, data.toByteArray(), Duration.ofMinutes(5));
        log.debug("Shortened URL: {} to code: {}", originalUrl, code);
        return code;
    }

    public String getOriginalUrl(String code) {
        byte[] bytes = redisTemplate.opsForValue().get(code);
        if (bytes == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.parseFrom(bytes);

            shortenedUrlUpdater.updateUrlStats(code, data);

            log.debug("Retrieved original URL: {} for code: {}", data.getOriginalUrl(), code);
            return data.getOriginalUrl();
        } catch (Exception e) {
            throw new UrlSerializationException(e);
        }
    }


    public ShortenedUrlStats getStats(String code) {
        log.debug("Retrieving stats for code: {}", code);
        byte[] bytes = redisTemplate.opsForValue().get(code);
        if (bytes == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.parseFrom(bytes);
            log.debug("Retrieved stats for code: {}", code);

            Instant createdAt = Instant.ofEpochMilli(data.getCreatedAt());
            Instant lastClickedAt = data.getLastClickedAt() == 0 ? null : Instant.ofEpochMilli(data.getLastClickedAt());

            return new ShortenedUrlStats(code,
                    data.getOriginalUrl(),
                    data.getClickCounter(),
                    createdAt,
                    lastClickedAt);
        } catch (Exception e) {
            throw new UrlSerializationException(e);
        }
    }
}