package urlshortener.url;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.ValueOperations;
import urlshortener.proto.ShortenedUrl;

import java.time.Duration;
import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.fail;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ShortenedUrlUpdaterTest {

    @Mock
    private RedisTemplate<String, byte[]> redisTemplate;

    @Mock
    private ValueOperations<String, byte[]> valueOperations;

    @InjectMocks
    private ShortenedUrlUpdater shortenedUrlUpdater;

    @Test
    void testUpdateUrlStats() {
        String code = "kVOkZ";
        ShortenedUrl.ShortenedUrlData data =
                ShortenedUrl.ShortenedUrlData.newBuilder()
                        .setOriginalUrl("https://example.com")
                        .setClickCounter(5)
                        .setCreatedAt(Instant.now().minusSeconds(10).toEpochMilli())
                        .setLastClickedAt(Instant.now().toEpochMilli())
                        .build();

        when(redisTemplate.opsForValue()).thenReturn(valueOperations);

        shortenedUrlUpdater.updateUrlStats(code, data);

        ArgumentCaptor<byte[]> dataCaptor = ArgumentCaptor.forClass(byte[].class);
        verify(valueOperations).set(eq(code), dataCaptor.capture(), eq(Duration.ofMinutes(5)));

        try {
            var updatedData = ShortenedUrl.ShortenedUrlData.parseFrom(dataCaptor.getValue());
            assertThat(updatedData.getClickCounter()).isEqualTo(6);
        } catch (Exception e) {
            fail("Failed to parse updated data");
        }
    }
}
