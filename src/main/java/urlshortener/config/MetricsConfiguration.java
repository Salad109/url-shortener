package urlshortener.config;

import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.core.StringRedisTemplate;

@Configuration
public class MetricsConfiguration {

    public MetricsConfiguration(MeterRegistry meterRegistry, StringRedisTemplate redisTemplate) {
        Gauge.builder("shortened_urls", () -> countUrlKeys(redisTemplate)).register(meterRegistry);
    }

    private int countUrlKeys(StringRedisTemplate redisTemplate) {
        try {
            return redisTemplate.keys("*").size() - 1; // 1 for url_counter key
        } catch (Exception e) {
            return 0;
        }
    }
}