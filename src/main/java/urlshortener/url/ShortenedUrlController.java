package urlshortener.url;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
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

    @Operation(summary = "Shorten a URL",
            description = "Takes a long URL and returns a short code related to it")
    @ApiResponse(responseCode = "201",
            description = "Short code created successfully",
            content = @Content(
                    mediaType = "text/plain",
                    examples = @ExampleObject(
                            name = "Successful shortening",
                            description = "Returns a 5-character alphanumeric code",
                            value = "kVOkZ"
                    )
            ))
    @ApiResponse(responseCode = "400",
            description = "Invalid URL provided",
            content = @Content(
                    mediaType = "application/json",
                    examples = @ExampleObject(
                            name = "Invalid URL error",
                            description = "Validation error response",
                            value = """
                                    {
                                      "originalUrl": "URL must start with http:// or https:// and be less than 2048 characters"
                                    }
                                    """
                    )
            ))
    @PostMapping("/shorten")
    @ResponseStatus(HttpStatus.CREATED)
    public String shortenUrl(@Valid @RequestBody ShortenRequest request) {
        return shortenedUrlService.shortenUrl(request.originalUrl());
    }

    @Operation(summary = "Redirect to original URL",
            description = "Redirects to the original URL based on the short code. Updates the short code's statistics.")
    @ApiResponse(responseCode = "302", description = "Redirected to original URL")
    @ApiResponse(responseCode = "404", description = "Short code not found, or invalid code")
    @Parameter(
            name = "code",
            description = "Short code representing the original URL",
            example = "kVOkZ"
    )
    @GetMapping("/{code}")
    public RedirectView redirect(@PathVariable @ValidCode String code) {
        String originalUrl = shortenedUrlService.getOriginalUrl(code);
        return new RedirectView(originalUrl);
    }

    @Operation(summary = "Get stats for a short code",
            description = "Returns statistics for the given short code, including creation date and click count")
    @ApiResponse(responseCode = "200", description = "Stats retrieved successfully")
    @ApiResponse(responseCode = "404", description = "Short code not found, or invalid code")
    @Parameter(
            name = "code",
            description = "Short code to retrieve statistics for",
            example = "kVOkZ"
    )
    @GetMapping("/stats/{code}")
    public ShortenedUrlStats getStats(@PathVariable @ValidCode String code) {
        return shortenedUrlService.getStats(code);
    }
}
