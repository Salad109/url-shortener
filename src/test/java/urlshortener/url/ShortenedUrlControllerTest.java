package urlshortener.url;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.test.web.servlet.assertj.MockMvcTester;
import org.springframework.test.web.servlet.assertj.MvcTestResult;
import urlshortener.dto.ShortenedUrlStats;

import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

@WebMvcTest(ShortenedUrlController.class)
class ShortenedUrlControllerTest {

    @Autowired
    private MockMvcTester mvc;

    @MockitoBean
    private ShortenedUrlService shortenedUrlService;

    @Test
    void testShortenUrl() {
        String expectedCode = "kVOkZ";
        when(shortenedUrlService.shortenUrl("https://example.com")).thenReturn(expectedCode);

        assertThat(mvc.post()
                .uri("/shorten")
                .contentType(MediaType.APPLICATION_JSON)
                .content("""
                        {"originalUrl": "https://example.com"}
                        """))
                .hasStatus(HttpStatus.CREATED)
                .hasBodyTextEqualTo(expectedCode);
    }

    @Test
    void testRedirect() {
        String originalUrl = "https://example.com";
        String code = "kVOkZ";
        when(shortenedUrlService.getOriginalUrl(code)).thenReturn(originalUrl);

        assertThat(mvc.get()
                .uri("/{code}", code)
                .accept(MediaType.TEXT_PLAIN))
                .hasStatus(HttpStatus.FOUND)
                .hasHeader("Location", originalUrl);
    }

    @Test
    void testGetStats() {
        String code = "kVOkZ";
        String originalUrl = "https://example.com";
        int clickCounter = 10;
        Instant createdAt = Instant.now().minusSeconds(10);
        Instant lastClickedAt = Instant.now();
        ShortenedUrlStats stats = new ShortenedUrlStats(
                code,
                originalUrl,
                clickCounter,
                createdAt,
                lastClickedAt
        );
        when(shortenedUrlService.getStats(code)).thenReturn(stats);

        MvcTestResult result = mvc.get().uri("/stats/{code}", code).accept(MediaType.APPLICATION_JSON).exchange();
        assertThat(result).hasStatus(HttpStatus.OK);
        assertThat(result).bodyJson().extractingPath("code").isEqualTo(code);
        assertThat(result).bodyJson().extractingPath("originalUrl").isEqualTo(originalUrl);
        assertThat(result).bodyJson().extractingPath("clickCounter").isEqualTo(clickCounter);
        assertThat(result).bodyJson().extractingPath("createdAt").isEqualTo(createdAt.toString());
        assertThat(result).bodyJson().extractingPath("lastClickedAt").isEqualTo(lastClickedAt.toString());
    }
}