package urlshortener.url;

import io.micrometer.core.instrument.MeterRegistry;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
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
}