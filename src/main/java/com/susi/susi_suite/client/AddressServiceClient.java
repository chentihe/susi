package com.susi.susi_suite.client;

import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.util.UriComponentsBuilder;

@Service
@RequiredArgsConstructor
public class AddressServiceClient {
    private final RestTemplate restTemplate;

    @Value("${address.service.url}")
    private String addressServiceUrl;

    public String getZipcodes(String road, Integer houseNumber) {

        // 使用 UriComponentsBuilder 處理中文路名編碼，避免亂碼
        String url = UriComponentsBuilder.fromHttpUrl(addressServiceUrl + "/zipcode")
                .queryParam("road", road)
                .queryParam("number", houseNumber)
                .toUriString();
        return restTemplate.getForObject(url, String.class);
    }
}
