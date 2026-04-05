package com.susi.susi_suite.dto.request;

import com.susi.susi_suite.entity.UnitStatus;

import java.util.List;
import java.util.Map;

public record UnitRequest(
        Long propertyId,
        String roomNumber,
        Integer rentAmount,
        Double squareFootage,
        UnitStatus status,
        Map<String, Object> interiorAmenities,
        List<String> images
) {
    public UnitRequest {
        if (rentAmount < 0) {
            throw new IllegalArgumentException("rent amount needs to be greater than 0");
        }
    }
}
