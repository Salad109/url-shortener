package urlshortener.url;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;
import urlshortener.transcoding.Base62Converter;
import urlshortener.transcoding.IdScrambler;

@Component
public class IdGenerator {

    private final StringRedisTemplate redisTemplate;

    public IdGenerator(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    public String generateCode() {
        Long id = redisTemplate.opsForValue().increment("internal:url_id_counter");
        if (id == null) {
            throw new IllegalStateException("ID increment failed");
        }
        long scrambledId = IdScrambler.encode(id);
        return Base62Converter.encode(scrambledId);
    }
}
