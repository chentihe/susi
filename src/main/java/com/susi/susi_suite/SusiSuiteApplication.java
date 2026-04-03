package com.susi.susi_suite;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.jpa.repository.config.EnableJpaAuditing;

@SpringBootApplication
@EnableJpaAuditing
public class SusiSuiteApplication {

	public static void main(String[] args) {
		SpringApplication.run(SusiSuiteApplication.class, args);
	}

}
