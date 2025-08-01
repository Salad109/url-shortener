package urlshortener.url;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.view.RedirectView;
import urlshortener.dto.ShortenRequest;
import urlshortener.dto.ShortenedUrlStats;

@RestController
public class ShortenedUrlController {

    private final ShortenedUrlService shortenedUrlService;

    public ShortenedUrlController(ShortenedUrlService shortenedUrlService) {
        this.shortenedUrlService = shortenedUrlService;
    }

    @PostMapping("/shorten")
    @ResponseStatus(HttpStatus.CREATED)
    public String shortenUrl(@Valid @RequestBody ShortenRequest request) {
        return shortenedUrlService.shortenUrl(request.originalUrl());
    }

    @GetMapping("/{code}")
    public RedirectView redirect(@PathVariable @NotBlank String code) {
        String originalUrl = shortenedUrlService.getOriginalUrl(code);
        return new RedirectView(originalUrl);
    }

    @GetMapping("/stats/{code}")
    public ShortenedUrlStats getShortenedUrlStats(@PathVariable String code) {
        return shortenedUrlService.getShortenedUrlStats(code);
    }
}
