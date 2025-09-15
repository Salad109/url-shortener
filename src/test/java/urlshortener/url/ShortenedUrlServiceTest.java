package urlshortener.url;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.ValueOperations;
import urlshortener.dto.ShortenResponse;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.proto.ShortenedUrl;

import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ShortenedUrlServiceTest {

    @Mock
    private RedisTemplate<String, byte[]> redisTemplate;

    @Mock
    private ValueOperations<String, byte[]> valueOperations;

    @Mock
    private IdGenerator idGenerator;

    @Mock
    private ShortenedUrlUpdater shortenedUrlUpdater;

    @InjectMocks
    private ShortenedUrlService shortenedUrlService;

    @Test
    void testShortenUrl() {
        String originalUrl = "http://example.com";
        String expectedCode = "12345";
        when(idGenerator.generateCode()).thenReturn(expectedCode);
        when(redisTemplate.opsForValue()).thenReturn(valueOperations);

        ShortenResponse result = shortenedUrlService.shortenUrl(originalUrl);
        String resultCode = result.shortCode();

        assertThat(resultCode).isEqualTo(expectedCode);
    }

    @Test
    void testGetOriginalUrl() {
        String code = "12345";
        String originalUrl = "http://example.com";
        Instant now = Instant.now();

        ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.newBuilder()
                .setOriginalUrl(originalUrl)
                .setClickCounter(5)
                .setCreatedAt(now.toEpochMilli())
                .setLastClickedAt(now.minusSeconds(100).toEpochMilli())
                .build();

        byte[] dataBytes = data.toByteArray();

        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.get(code)).thenReturn(dataBytes);

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
    void testGetStats() {
        String code = "12345";
        String originalUrl = "http://example.com";
        long createdAt = Instant.now().minusSeconds(10).toEpochMilli();
        long lastClickedAt = Instant.now().toEpochMilli();
        long clickCounter = 21;

        ShortenedUrl.ShortenedUrlData data = ShortenedUrl.ShortenedUrlData.newBuilder()
                .setOriginalUrl(originalUrl)
                .setClickCounter(clickCounter)
                .setCreatedAt(createdAt)
                .setLastClickedAt(lastClickedAt)
                .build();

        byte[] dataBytes = data.toByteArray();

        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.get(code)).thenReturn(dataBytes);

        ShortenedUrlStats result = shortenedUrlService.getStats(code);

        assertThat(result.code()).isEqualTo(code);
        assertThat(result.originalUrl()).isEqualTo(originalUrl);
        assertThat(result.clickCounter()).isEqualTo(clickCounter);
        assertThat(result.createdAt().toEpochMilli()).isEqualTo(createdAt);
        assertThat(result.lastClickedAt().toEpochMilli()).isEqualTo(lastClickedAt);
    }
}