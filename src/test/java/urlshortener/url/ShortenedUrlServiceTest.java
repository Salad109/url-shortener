package urlshortener.url;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.exception.UrlSerializationException;

import java.time.Instant;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ShortenedUrlServiceTest {

    @Mock
    private StringRedisTemplate redisTemplate;

    @Mock
    private ValueOperations<String, String> valueOperations;

    @Mock
    private IdGenerator idGenerator;

    private ShortenedUrlService shortenedUrlService;

    @BeforeEach
    void setUp() {
        ObjectMapper objectMapper = new ObjectMapper().registerModule(new JavaTimeModule());
        shortenedUrlService = new ShortenedUrlService(redisTemplate, idGenerator, objectMapper);
    }

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
    void testShortenUrlException() throws Exception {
        ObjectMapper mockedObjectMapper = mock(ObjectMapper.class);
        ShortenedUrlService mockedService = new ShortenedUrlService(redisTemplate, idGenerator, mockedObjectMapper);

        when(mockedObjectMapper.writeValueAsString(any(urlshortener.url.ShortenedUrlData.class)))
                .thenThrow(new JsonProcessingException("Serialization error") {
                });
        when(idGenerator.generateCode()).thenReturn("12345");


        assertThatThrownBy(() -> mockedService.shortenUrl("http://example.com"))
                .isInstanceOf(UrlSerializationException.class)
                .hasMessage("Error serializing or deserializing URL data");
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