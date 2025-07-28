package urlshortener.dto;

import jakarta.validation.constraints.NotBlank;

public record ShortenRequest(@NotBlank String originalUrl) {

}
