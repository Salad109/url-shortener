package urlshortener.url;

import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;
import urlshortener.dto.ShortenRequest;

@RestController
public class ShortenedUrlController {

    private final ShortenedUrlService shortenedUrlService;

    public ShortenedUrlController(ShortenedUrlService shortenedUrlService) {
        this.shortenedUrlService = shortenedUrlService;
    }

    @PostMapping("/shorten")
    public String shortenUrl(@Valid @RequestBody ShortenRequest request) {
        return shortenedUrlService.shortenUrl(request.originalUrl());
    }

    @GetMapping("/{code}")
    public String getOriginalUrl(@PathVariable String code) {
        return shortenedUrlService.getOriginalUrl(code);
    }
}
