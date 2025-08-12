package urlshortener.config;

import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.RedisServerCommands;
import org.springframework.data.redis.core.StringRedisTemplate;

@Configuration
public class MetricsConfiguration {

    public MetricsConfiguration(MeterRegistry meterRegistry, StringRedisTemplate redisTemplate) {
        Gauge.builder("shortened_urls", () -> countUrlKeys(redisTemplate))
                .description("Number of shortened URLs currently stored")
                .register(meterRegistry);
    }

    private long countUrlKeys(StringRedisTemplate redisTemplate) {
        try {
            Long totalKeys = redisTemplate.execute(RedisServerCommands::dbSize);

            return totalKeys == null ? 0 : Math.max(0, totalKeys - 1); // -1 for the id counter key
        } catch (Exception e) {
            return 0;
        }
    }
}