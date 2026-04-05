package com.susi.susi_suite.dto.response;

import java.util.List;
import java.util.Map;

public record PropertyResponse(
        Long id,
        String name,
        String description,
        String city,
        String area,
        String road,
        String scope,
        String address,
        String zipcode,
        Integer totalFloors,
        Integer floor,
        Integer basementFloors,
        Map<String, Object> publicAmenities,
        List<UnitResponse> units,
        String coverImage,
        List<String> images
) {
}
