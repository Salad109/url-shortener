package urlshortener.url;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.exception.UrlDeserializationException;
import urlshortener.exception.UrlSerializationException;

import java.time.Duration;

@Service
public class ShortenedUrlService {

    private final StringRedisTemplate redisTemplate;
    private final IdGenerator idGenerator;
    private final ObjectMapper objectMapper;
    private static final Logger log = LoggerFactory.getLogger(ShortenedUrlService.class);

    public ShortenedUrlService(StringRedisTemplate redisTemplate, IdGenerator idGenerator, ObjectMapper objectMapper) {
        this.redisTemplate = redisTemplate;
        this.idGenerator = idGenerator;
        this.objectMapper = objectMapper;
    }

    public String shortenUrl(String originalUrl) {
        String code = idGenerator.generateCode();
        ShortenedUrlData shortenedUrlData = new ShortenedUrlData(originalUrl);

        try {
            log.info("Shortening URL: {} with code: {}", originalUrl, code);
            String jsonData = objectMapper.writeValueAsString(shortenedUrlData);
            redisTemplate.opsForValue().set(code, jsonData, Duration.ofMinutes(5));
            log.debug("Shortened URL: {} to code: {}", originalUrl, code);
        } catch (JsonProcessingException e) {
            throw new UrlSerializationException(e);
        }

        return code;
    }

    public String getOriginalUrl(String code) {
        log.debug("Retrieving original URL for code: {}", code);
        String jsonData = redisTemplate.opsForValue().get(code);
        if (jsonData == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrlData shortenedUrlData = objectMapper.readValue(jsonData, ShortenedUrlData.class);
            shortenedUrlData = shortenedUrlData.incrementClicks();
            jsonData = objectMapper.writeValueAsString(shortenedUrlData);
            redisTemplate.opsForValue().set(code, jsonData, Duration.ofMinutes(5));
            log.debug("Retrieved original URL: {} for code: {}", shortenedUrlData.originalUrl(), code);

            return shortenedUrlData.originalUrl();
        } catch (JsonProcessingException e) {
            throw new UrlDeserializationException(e);
        }
    }

    public ShortenedUrlStats getStats(String code) {
        log.debug("Retrieving stats for code: {}", code);
        String jsonData = redisTemplate.opsForValue().get(code);
        if (jsonData == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrlData shortenedUrlData = objectMapper.readValue(jsonData, ShortenedUrlData.class);
            log.debug("Retrieved stats for code: {}", code);

            return new ShortenedUrlStats(code,
                    shortenedUrlData.originalUrl(),
                    shortenedUrlData.clickCounter(),
                    shortenedUrlData.createdAt(),
                    shortenedUrlData.lastClickedAt());
        } catch (JsonProcessingException e) {
            throw new UrlDeserializationException(e);
        }
    }
}