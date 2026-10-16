package com.sundial;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;

@SpringBootApplication
@ConfigurationPropertiesScan
public class SundialApplication {
    public static void main(String[] args) {
        SpringApplication.run(SundialApplication.class, args);
    }
}
