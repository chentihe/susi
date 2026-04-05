package com.susi.susi_suite.controller;

import com.susi.susi_suite.dto.request.UnitRequest;
import com.susi.susi_suite.dto.response.UnitResponse;
import com.susi.susi_suite.entity.Unit;
import com.susi.susi_suite.service.UnitService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/units")
@RequiredArgsConstructor
@Slf4j
public class UnitController {
    private final UnitService unitService;

    @GetMapping("/{id}")
    public ResponseEntity<UnitResponse> getUnitById(@PathVariable Long id) {
        return ResponseEntity.ok(unitService.getUnitById(id));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN')")
    public ResponseEntity<UnitResponse> addUnit(@RequestBody UnitRequest request) {
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(unitService.createUnit(request));
    }
}
