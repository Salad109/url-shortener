package urlshortener.url;

import io.micrometer.core.instrument.MeterRegistry;
import jakarta.persistence.EntityNotFoundException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import urlshortener.dto.ShortenedUrlStats;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ShortenedUrlServiceTest {

    @Mock
    private ShortenedUrlRepository repository;

    @Mock
    private MeterRegistry meterRegistry;

    @InjectMocks
    private ShortenedUrlService service;

    @Test
    void testShortenUrl() {
        String originalUrl = "https://example.com";

        ShortenedUrl savedUrl = new ShortenedUrl();
        savedUrl.setOriginalUrl(originalUrl);
        savedUrl.setId(1L);

        when(repository.save(any(ShortenedUrl.class))).thenReturn(savedUrl);

        String result = service.shortenUrl(originalUrl);

        assertThat(result).isEqualTo("kVOkZ"); // 1 -> 687194767 -> kVOkZ
    }

    @Test
    void testGetOriginalUrl() {
        String originalUrl = "https://example.com";

        ShortenedUrl url = new ShortenedUrl();
        url.setId(1L);
        url.setOriginalUrl(originalUrl);

        when(repository.findById(1L)).thenReturn(java.util.Optional.of(url));

        String result = service.getOriginalUrl("kVOkZ");

        assertThat(result).isEqualTo(originalUrl);
    }

    @Test
    void testGetOriginalUrlNotFound() {
        when(repository.findById(1L)).thenReturn(java.util.Optional.empty());

        assertThatThrownBy(() -> service.getOriginalUrl("kVOkZ"))
                .isInstanceOf(EntityNotFoundException.class);

    }

    @Test
    void testGetOriginalUrlTooLong() {
        assertThatThrownBy(() -> service.getOriginalUrl("1234567890"))
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void testGetShortenedUrlStats() {
        String originalUrl = "https://example.com";

        ShortenedUrl url = new ShortenedUrl();
        url.setId(1L);
        url.setOriginalUrl(originalUrl);

        when(repository.findById(1L)).thenReturn(java.util.Optional.of(url));

        ShortenedUrlStats stats = service.getShortenedUrlStats("kVOkZ");

        assertThat(stats.code()).isEqualTo("kVOkZ");
        assertThat(stats.originalUrl()).isEqualTo(originalUrl);
        assertThat(stats.clickCounter()).isZero();
        assertThat(stats.createdAt()).isNotNull();
        assertThat(stats.lastClickedAt()).isNull();
    }
}