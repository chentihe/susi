package com.susi.susi_suite.mapper;

import com.susi.susi_suite.dto.request.PropertyReqeust;
import com.susi.susi_suite.dto.response.PropertyResponse;
import com.susi.susi_suite.dto.response.UnitResponse;
import com.susi.susi_suite.entity.Property;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Component
@RequiredArgsConstructor
public class PropertyConverter {
    private final UnitConverter unitConverter;

    public Property toEntity(PropertyReqeust request) {
        if (request == null) {
            return null;
        }

        Property property = new Property();
        property.setName(request.name());
        property.setDescription(request.description());
        property.setCity(request.city());
        property.setArea(request.area());
        property.setRoad(request.road());
        property.setScope(request.scope());
        property.setZipcode(request.zipcode());
        property.setTotalFloors(request.totalFloors());
        property.setFloor(request.floor());
        property.setBasementFloors(request.basementFloors());
        property.setLandlordId(request.landlordId());
        property.setPublicAmenities(request.publicAmenities());
        property.setCoverImage(request.coverImage());
        property.setImages(request.images());
        return property;
    }

    public PropertyResponse toResponse(Property property) {
        if (property == null) {
            return null;
        }

        List<UnitResponse> unitResponseStream = property.getUnits()
                .stream().map(u -> unitConverter.toResponse(u))
                .collect(Collectors.toList());

        return new PropertyResponse(
                property.getId(),
                property.getName(),
                property.getDescription(),
                property.getCity(),
                property.getArea(),
                property.getRoad(),
                property.getScope(),
                property.getAddress(),
                property.getZipcode(),
                property.getTotalFloors(),
                property.getFloor(),
                property.getBasementFloors(),
                property.getPublicAmenities(),
                unitResponseStream,
                property.getCoverImage(),
                property.getImages()
        );
    }
}
