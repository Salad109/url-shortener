package urlshortener.exception;

public class UrlDeserializationException extends RuntimeException {
    public UrlDeserializationException(Throwable cause) {
        super("Error deserializing URL data", cause);
    }
}
