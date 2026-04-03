package com.susi.susi_suite.controller;

import lombok.extern.slf4j.Slf4j;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.core.user.OAuth2User;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api/v1/user")
@Slf4j
public class UserController {

    @GetMapping("/me")
    public Map<String, Object> getUser(@AuthenticationPrincipal OAuth2User principal) {
        log.info("OAuth2 info: {}", principal.toString());
        return principal.getAttributes();
    }
}
