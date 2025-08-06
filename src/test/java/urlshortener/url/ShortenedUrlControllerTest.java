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
    private MockMvcTester mockMvcTester;

    @MockitoBean
    private ShortenedUrlService shortenedUrlService;

    @Test
    void shouldShortenUrl() {
        String originalUrl = "https://example.com";
        String expectedCode = "abcd";
        when(shortenedUrlService.shortenUrl(originalUrl)).thenReturn(expectedCode);

        assertThat(mockMvcTester.post().uri("/shorten")
                .contentType(MediaType.APPLICATION_JSON)
                .content("""
                        {
                            "originalUrl": "https://example.com"
                        }
                        """)).hasStatus(HttpStatus.CREATED)
                .bodyText()
                .isEqualTo(expectedCode);
    }

    @Test
    void shouldRejectInvalidUrl() {
        assertThat(mockMvcTester.post().uri("/shorten")
                .contentType(MediaType.APPLICATION_JSON)
                .content("""
                        {
                            "originalUrl": "goober"
                        }
                        """)).hasStatus(HttpStatus.BAD_REQUEST)
                .bodyJson()
                .extractingPath("originalUrl")
                .asString()
                .contains("URL must start with http");
    }

    @Test
    void testRedirect() {
        String code = "abcd";
        String originalUrl = "https://example.com";
        when(shortenedUrlService.getOriginalUrl(code)).thenReturn(originalUrl);

        assertThat(mockMvcTester.get().uri("/{code}", code))
                .hasStatus(HttpStatus.FOUND)
                .hasHeader("Location", originalUrl);
    }

    @Test
    void testGetShortenedUrlStats() {
        String code = "abcd";
        String originalUrl = "https://example.com";
        int clickCounter = 10;
        ShortenedUrlStats stats = new ShortenedUrlStats(code, originalUrl, clickCounter, Instant.now(), Instant.now());
        when(shortenedUrlService.getShortenedUrlStats(code)).thenReturn(stats);

        MvcTestResult result = mockMvcTester.get().uri("/stats/{code}", code).exchange();

        assertThat(result).hasStatus(HttpStatus.OK);
        assertThat(result).bodyJson().extractingPath("code").asString().isEqualTo(code);
        assertThat(result).bodyJson().extractingPath("originalUrl").asString().isEqualTo(originalUrl);
        assertThat(result).bodyJson().extractingPath("clickCounter").asNumber().isEqualTo(clickCounter);
        assertThat(result).bodyJson().extractingPath("createdAt").isNotNull();
        assertThat(result).bodyJson().extractingPath("lastClickedAt").isNotNull();
    }
}