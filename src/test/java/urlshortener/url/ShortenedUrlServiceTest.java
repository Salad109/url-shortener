package urlshortener.url;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;
import urlshortener.dto.ShortenedUrlStats;

import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ShortenedUrlServiceTest {

    @Mock
    private StringRedisTemplate redisTemplate;

    @Mock
    private ValueOperations<String, String> valueOperations;

    @Mock
    private IdGenerator idGenerator;

    @InjectMocks
    private ShortenedUrlService shortenedUrlService;


    @Test
    void testShortenUrl() {
        String originalUrl = "http://example.com";
        String expectedCode = "12345";
        when(idGenerator.generateCode()).thenReturn(expectedCode);
        when(redisTemplate.opsForValue()).thenReturn(valueOperations);

        String resultCode = shortenedUrlService.shortenUrl(originalUrl);

        assertThat(expectedCode).isEqualTo(resultCode);
    }

    @Test
    void testGetOriginalUrl() throws Exception {
        String code = "12345";
        String originalUrl = "http://example.com";
        urlshortener.url.ShortenedUrlData shortenedUrlData = new urlshortener.url.ShortenedUrlData(originalUrl);
        ObjectMapper objectMapper = new ObjectMapper().registerModule(new JavaTimeModule());
        String jsonData = objectMapper.writeValueAsString(shortenedUrlData);

        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.get(code)).thenReturn(jsonData);

        String resultUrl = shortenedUrlService.getOriginalUrl(code);

        assertThat(resultUrl).isEqualTo(originalUrl);
    }

    @Test
    void testGetOriginalUrlNotFound() {
        String code = "12345";
        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.get(code)).thenReturn(null);

        assertThatThrownBy(() -> shortenedUrlService.getOriginalUrl(code))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("Short URL not found");
    }

    @Test
    void testGetStats() throws Exception {
        String code = "12345";
        Instant now = Instant.now();
        ShortenedUrlData data = new ShortenedUrlData("http://example.com", 10, now, now);
        ObjectMapper objectMapper = new ObjectMapper().registerModule(new JavaTimeModule());

        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.get(code)).thenReturn(objectMapper.writeValueAsString(data));

        ShortenedUrlStats result = shortenedUrlService.getStats(code);

        assertThat(result.code()).isEqualTo(code);
        assertThat(result.originalUrl()).isEqualTo(data.originalUrl());
        assertThat(result.clickCounter()).isEqualTo(data.clickCounter());
        assertThat(result.createdAt()).isEqualTo(data.createdAt());
        assertThat(result.lastClickedAt()).isEqualTo(data.lastClickedAt());
    }
}