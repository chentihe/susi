package com.susi.susi_suite.client;

import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

@Service
@RequiredArgsConstructor
public class AddressServiceClient {
    private final RestTemplate restTemplate;

    @Value("${address.service.url}")
    private String addressServiceUrl;

    public String getZipcodes(String road, Integer houseNumber) {
        String url = String.format("%s/zipcode?road=%s&number=%s", addressServiceUrl, road, houseNumber);
        return restTemplate.getForObject(url, String.class);
    }
}
