package urlshortener.url;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.ValueOperations;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class IdGeneratorTest {

    @Mock
    private RedisTemplate<String, byte[]> redisTemplate;

    @Mock
    private ValueOperations<String, byte[]> valueOperations;

    @InjectMocks
    private IdGenerator idGenerator;

    @Test
    void testGenerateCode() {
        long id = 1;
        String code = "kVOkZ";
        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.increment("internal:url_id_counter")).thenReturn(id);

        assertThat(idGenerator.generateCode()).isEqualTo(code);
    }

    @Test
    void testGenerateCodeNonExistentId() {
        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        when(valueOperations.increment("internal:url_id_counter")).thenReturn(null);

        assertThatThrownBy(() -> idGenerator.generateCode())
                .isInstanceOf(IllegalStateException.class)
                .hasMessage("ID increment failed");
    }
}
