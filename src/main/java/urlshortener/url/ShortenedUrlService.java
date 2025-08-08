package urlshortener.url;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
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

    public ShortenedUrlService(StringRedisTemplate redisTemplate, IdGenerator idGenerator, ObjectMapper objectMapper) {
        this.redisTemplate = redisTemplate;
        this.idGenerator = idGenerator;
        this.objectMapper = objectMapper;
    }

    public String shortenUrl(String originalUrl) {
        String code = idGenerator.generateCode();
        ShortenedUrlData shortenedUrlData = new ShortenedUrlData(originalUrl);

        try {
            String jsonData = objectMapper.writeValueAsString(shortenedUrlData);
            redisTemplate.opsForValue().set(code, jsonData, Duration.ofMinutes(5));
        } catch (JsonProcessingException e) {
            throw new UrlSerializationException(e);
        }

        return code;
    }

    public String getOriginalUrl(String code) {
        String jsonData = redisTemplate.opsForValue().get(code);
        if (jsonData == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrlData shortenedUrlData = objectMapper.readValue(jsonData, ShortenedUrlData.class);
            shortenedUrlData = shortenedUrlData.incrementClicks();
            jsonData = objectMapper.writeValueAsString(shortenedUrlData);
            redisTemplate.opsForValue().set(code, jsonData, Duration.ofMinutes(5));

            return shortenedUrlData.originalUrl();
        } catch (JsonProcessingException e) {
            throw new UrlDeserializationException(e);
        }
    }

    public ShortenedUrlStats getStats(String code) {
        String jsonData = redisTemplate.opsForValue().get(code);
        if (jsonData == null) {
            throw new IllegalArgumentException("Short URL not found");
        }

        try {
            ShortenedUrlData shortenedUrlData = objectMapper.readValue(jsonData, ShortenedUrlData.class);

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