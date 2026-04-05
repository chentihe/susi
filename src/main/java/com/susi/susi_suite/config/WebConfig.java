package com.susi.susi_suite.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class WebConfig implements WebMvcConfigurer {

    @Value("${app.upload-path:/app/uploads}")
    private String uploadPath;

    @Override
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        // 確保路徑以 file: 開頭，且最後面有斜線
        String location = "file:" + uploadPath + "/";

        registry.addResourceHandler("/images/**")
                .addResourceLocations(location);
    }
}
