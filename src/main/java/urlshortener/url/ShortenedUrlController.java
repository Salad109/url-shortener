package urlshortener.url;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.ExampleObject;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.view.RedirectView;
import urlshortener.dto.ShortenRequest;
import urlshortener.dto.ShortenResponse;
import urlshortener.dto.ShortenedUrlStats;
import urlshortener.validation.ValidCode;

@Tag(name = "URL Shortener", description = "API for shortening URLs, redirecting and retrieving statistics")
@Validated
@RestController
public class ShortenedUrlController {

    private final ShortenedUrlService shortenedUrlService;

    public ShortenedUrlController(ShortenedUrlService shortenedUrlService) {
        this.shortenedUrlService = shortenedUrlService;
    }

    @Operation(summary = "Home page")
    @ApiResponse(responseCode = "302", description = "Redirected to homepage")
    @GetMapping("/")
    public RedirectView home() {
        return new RedirectView("/static/index.html");
    }

    @Operation(summary = "Shorten a URL",
            description = "Takes a long URL and returns a short code related to it")
    @ApiResponse(responseCode = "201", description = "Short code created successfully")
    @ApiResponse(responseCode = "400", description = "Invalid URL provided",
            content = @Content(examples = @ExampleObject(value = "{\"originalUrl\": \"Invalid URL\"}")))
    @PostMapping("/shorten")
    @ResponseStatus(HttpStatus.CREATED)
    public ShortenResponse shortenUrl(@Valid @RequestBody ShortenRequest request) {
        return shortenedUrlService.shortenUrl(request.originalUrl());
    }

    @Operation(summary = "Redirect to original URL")
    @ApiResponse(responseCode = "302", description = "Redirected to original URL")
    @ApiResponse(responseCode = "404", description = "Short code not found",
            content = @Content(examples = @ExampleObject(value = "{\"error\": \"Short URL not found\"}")))
    @GetMapping("/{code}")
    public RedirectView redirect(@PathVariable @ValidCode String code) {
        String originalUrl = shortenedUrlService.getOriginalUrl(code);
        return new RedirectView(originalUrl);
    }

    @Operation(summary = "Get stats for a short code")
    @ApiResponse(responseCode = "200", description = "Stats retrieved successfully")
    @ApiResponse(responseCode = "404", description = "Short code not found",
            content = @Content(examples = @ExampleObject(value = "{\"error\": \"Short URL not found\"}")))
    @GetMapping("/stats/{code}")
    public ShortenedUrlStats getStats(@PathVariable @ValidCode String code) {
        return shortenedUrlService.getStats(code);
    }
}