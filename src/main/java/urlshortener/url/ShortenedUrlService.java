package urlshortener.url;

import io.micrometer.core.instrument.MeterRegistry;
import jakarta.persistence.EntityNotFoundException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.repository.CrudRepository;
import org.springframework.stereotype.Service;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.transcoding.Base62Converter;
import urlshortener.transcoding.IdScrambler;

import java.time.Instant;
import java.util.Optional;

@Service
public class ShortenedUrlService {

    private static final Logger log = LoggerFactory.getLogger(ShortenedUrlService.class);
    ShortenedUrlRepository shortenedUrlRepository;

    ShortenedUrlService(ShortenedUrlRepository shortenedUrlRepository, MeterRegistry meterRegistry) {
        this.shortenedUrlRepository = shortenedUrlRepository;

        meterRegistry.gauge("shortened_urls_total", shortenedUrlRepository, CrudRepository::count);
    }

    public String shortenUrl(String originalUrl) {
        log.debug("Shortening URL: {}", originalUrl);
        ShortenedUrl shortenedUrl = new ShortenedUrl();
        shortenedUrl.setOriginalUrl(originalUrl);
        ShortenedUrl shortenedUrlEntity = shortenedUrlRepository.save(shortenedUrl);

        long scrambledId = IdScrambler.encode(shortenedUrlEntity.getId());
        String encodedId = Base62Converter.encode(scrambledId);
        log.info("Shortened URL created. ID: {}, code: {}, URL: {}", shortenedUrlEntity.getId(), encodedId, originalUrl);
        return encodedId;
    }

    public String getOriginalUrl(String code) {
        log.debug("Retrieving original URL for code: {}", code);
        ShortenedUrl url = getShortUrlOrThrow(code);

        url.incrementClickCounter();
        url.setLastClickedAt(Instant.now());

        shortenedUrlRepository.save(url);

        log.info("Original URL retrieved for code: {}, URL: {}", code, url.getOriginalUrl());
        return url.getOriginalUrl();
    }

    public ShortenedUrlStats getShortenedUrlStats(String code) {
        log.debug("Retrieving stats for code: {}", code);
        ShortenedUrl url = getShortUrlOrThrow(code);
        log.info("Stats retrieved for ID: {}, code: {}", url.getId(), code);
        return new ShortenedUrlStats(code, url.getOriginalUrl(), url.getClickCounter(), url.getCreatedAt(), url.getLastClickedAt());
    }

    private ShortenedUrl getShortUrlOrThrow(String code) {
        if (code.length() > 5) {
            log.debug("Code {} has invalid length: {}", code, code.length());
            throw new IllegalArgumentException("Short URL code must be 5 characters long");
        }

        long decodedId = Base62Converter.decode(code);
        long originalId = IdScrambler.decode(decodedId);

        Optional<ShortenedUrl> shortenedUrl = shortenedUrlRepository.findById(originalId);
        if (shortenedUrl.isEmpty()) {
            log.debug("Short URL not found for code: {}, ID: {}", code, originalId);
            throw new EntityNotFoundException();
        }

        ShortenedUrl foundUrl = shortenedUrl.get();
        log.debug("Short URL found for code: {}, ID: {}, Original URL: {}", code, originalId, foundUrl.getOriginalUrl());
        return foundUrl;
    }
}
