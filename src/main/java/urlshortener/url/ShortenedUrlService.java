package urlshortener.url;

import jakarta.persistence.EntityNotFoundException;
import org.springframework.stereotype.Service;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.transcoding.Base62Converter;
import urlshortener.transcoding.IdScrambler;

import java.time.Instant;
import java.util.Optional;

@Service
public class ShortenedUrlService {

    ShortenedUrlRepository shortenedUrlRepository;

    ShortenedUrlService(ShortenedUrlRepository shortenedUrlRepository) {
        this.shortenedUrlRepository = shortenedUrlRepository;
    }

    public String shortenUrl(String originalUrl) {
        ShortenedUrl shortenedUrl = new ShortenedUrl(originalUrl);
        ShortenedUrl shortenedUrlEntity = shortenedUrlRepository.save(shortenedUrl);

        long scrambledId = IdScrambler.encode(shortenedUrlEntity.getId());
        return Base62Converter.encode(scrambledId);
    }

    public String getOriginalUrl(String code) {
        ShortenedUrl url = getShortUrlOrThrow(code);

        url.incrementClickCounter();
        url.setLastClickedAt(Instant.now());

        shortenedUrlRepository.save(url);

        return url.getOriginalUrl();
    }

    public ShortenedUrlStats getShortenedUrlStats(String code) {
        ShortenedUrl url = getShortUrlOrThrow(code);
        return new ShortenedUrlStats(code, url.getOriginalUrl(), url.getClickCounter(), url.getCreatedAt(), url.getLastClickedAt());
    }

    private ShortenedUrl getShortUrlOrThrow(String code) {
        if (code.length() > 5) {
            throw new EntityNotFoundException("Short URL not found");
        }

        long decodedId = Base62Converter.decode(code);
        long originalId = IdScrambler.decode(decodedId);

        Optional<ShortenedUrl> shortenedUrl = shortenedUrlRepository.findById(originalId);
        if (shortenedUrl.isEmpty()) {
            throw new EntityNotFoundException("Short URL not found");
        }
        return shortenedUrl.get();
    }
}
