package com.davyj0nes.otel

import io.opentelemetry.api.metrics.DoubleCounter
import io.opentelemetry.api.metrics.GlobalMeterProvider
import io.opentelemetry.api.metrics.Meter
import io.opentelemetry.api.metrics.common.Labels
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import java.util.concurrent.atomic.AtomicLong


@SpringBootApplication
class OtelApplication

fun main(args: Array<String>) {
	runApplication<OtelApplication>(*args)
}

data class Greeting(val id: Long, val content: String)

@RestController
class IndexController {
	val counter = AtomicLong()
    private final val meter: Meter = GlobalMeterProvider.getMeter("opentelemetry-javaagent")
	val c: DoubleCounter = meter.doubleCounterBuilder("http_request_count_total").build()

	@GetMapping("/")
	fun index(@RequestParam(value = "name", defaultValue = "World") name: String): Greeting {
		c.add(1.0, Labels.of("service_name", "service_three", "path", "/"))
		return Greeting(counter.incrementAndGet(), "Hello, $name")
	}
}