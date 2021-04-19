package com.izettle.otel

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import java.util.concurrent.atomic.AtomicLong
import io.opentelemetry.api.metrics.GlobalMetricsProvider
import io.opentelemetry.api.metrics.LongCounter
import io.opentelemetry.api.metrics.Meter
import io.opentelemetry.api.metrics.MeterProvider

@SpringBootApplication
class OtelApplication

fun main(args: Array<String>) {
	runApplication<OtelApplication>(*args)
}

data class Greeting(val id: Long, val content: String)

@RestController
class IndexController {
	val counter = AtomicLong()
	private final val provider: MeterProvider = GlobalMetricsProvider.get()
    val meter: Meter = provider.get("opentelemetry-java-instrumentation")

	@GetMapping("/")
	fun index(@RequestParam(value = "name", defaultValue = "World") name: String): Greeting {
        val c: LongCounter = meter
				.longCounterBuilder("basic_counter")
				.setDescription("basic counter")
				.setUnit("1")
				.build()

		c.add(1)
		return Greeting(counter.incrementAndGet(), "Hello, $name")
	}

}