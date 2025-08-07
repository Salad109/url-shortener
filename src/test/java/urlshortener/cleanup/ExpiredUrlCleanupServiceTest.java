package urlshortener.cleanup;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import urlshortener.url.ShortenedUrlRepository;

import java.time.Instant;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ExpiredUrlCleanupServiceTest {

    @Mock
    private ShortenedUrlRepository repository;

    @InjectMocks
    private ExpiredUrlCleanupService cleanupService;

    @Test
    void shouldCallRepositoryToDeleteExpiredUrls() {
        when(repository.deleteExpiredUrls(any(Instant.class))).thenReturn(5);

        cleanupService.cleanupExpiredUrls();

        verify(repository).deleteExpiredUrls(any(Instant.class));
    }
}
