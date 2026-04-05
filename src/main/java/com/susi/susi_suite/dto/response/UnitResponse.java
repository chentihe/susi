package com.susi.susi_suite.dto.response;

import com.susi.susi_suite.entity.UnitStatus;

import java.util.List;
import java.util.Map;

public record UnitResponse(
        Long id,
        Long propertyId,
        String roomNumber,
        Integer rentAmount,
        Double squareFootage,
        UnitStatus status,
        Map<String, Object> interiorAmenities,
        List<String> images
) {
}
