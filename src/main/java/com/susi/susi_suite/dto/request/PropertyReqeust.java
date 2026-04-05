package com.susi.susi_suite.dto.request;

import java.util.List;
import java.util.Map;

public record PropertyReqeust(
        String name,
        String description,
        String city,
        String area,
        String road,
        String scope,
        String zipcode,
        Integer totalFloors,
        Integer floor,
        Integer basementFloors,
        Long landlordId,
        Map<String, Object> publicAmenities,
        String coverImage,
        List<String> images
) {
}
