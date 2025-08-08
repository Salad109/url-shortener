package urlshortener.exception;

public class UrlSerializationException extends RuntimeException {
    public UrlSerializationException(Throwable cause) {
        super("Error serializing or deserializing URL data", cause);
    }
}
