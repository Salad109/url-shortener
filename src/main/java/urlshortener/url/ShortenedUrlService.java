package urlshortener.url;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import urlshortener.dto.ShortenResponse;
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
    private final String baseUrl;

    public ShortenedUrlService(RedisTemplate<String, byte[]> redisTemplate, IdGenerator idGenerator, ShortenedUrlUpdater shortenedUrlUpdater, @Value("${app.base-url}") String baseUrl) {
        this.redisTemplate = redisTemplate;
        this.idGenerator = idGenerator;
        this.shortenedUrlUpdater = shortenedUrlUpdater;
        this.baseUrl = baseUrl;
    }

    public ShortenResponse shortenUrl(String originalUrl) {
        log.debug("Received request to shorten URL: {}", originalUrl);
        String code = idGenerator.generateCode();

        ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.newBuilder()
                .setOriginalUrl(originalUrl)
                .setClickCounter(0)
                .setCreatedAt(Instant.now().toEpochMilli())
                .setLastClickedAt(0)
                .build();

        redisTemplate.opsForValue().set(code, data.toByteArray(), Duration.ofMinutes(5));
        log.info("Shortened URL: {} to code: {}", originalUrl, code);
        String shortUrl = baseUrl + "/" + code;
        return new ShortenResponse(code, shortUrl);
    }

    public String getOriginalUrl(String code) {
        log.debug("Retrieving original URL for code: {}", code);
        byte[] bytes = redisTemplate.opsForValue().get(code);
        if (bytes == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.parseFrom(bytes);

            shortenedUrlUpdater.updateUrlStats(code, data);

            log.info("Retrieved original URL: {} for code: {}", data.getOriginalUrl(), code);
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

            Instant createdAt = Instant.ofEpochMilli(data.getCreatedAt());
            Instant lastClickedAt = data.getLastClickedAt() == 0 ? null : Instant.ofEpochMilli(data.getLastClickedAt());

            log.info("Retrieved stats for code: {}", code);

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