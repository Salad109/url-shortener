package urlshortener.url;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.time.Duration;

@Service
public class ShortenedUrlService {

    private final StringRedisTemplate redisTemplate;
    private final IdGenerator idGenerator;

    public ShortenedUrlService(StringRedisTemplate redisTemplate, IdGenerator idGenerator) {
        this.redisTemplate = redisTemplate;
        this.idGenerator = idGenerator;
    }

    public String shortenUrl(String originalUrl) {
        String code = idGenerator.generateCode();
        redisTemplate.opsForValue().set(code, originalUrl, Duration.ofMinutes(5));
        return code;
    }

    public String getOriginalUrl(String code) {
        String url = redisTemplate.opsForValue().get(code);
        if (url == null) {
            throw new IllegalArgumentException("Short URL not found");
        }
        return url;
    }
}