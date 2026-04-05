package com.susi.susi_suite.mapper;

import com.susi.susi_suite.dto.request.PropertyReqeust;
import com.susi.susi_suite.dto.request.UnitRequest;
import com.susi.susi_suite.dto.response.PropertyResponse;
import com.susi.susi_suite.dto.response.UnitResponse;
import com.susi.susi_suite.entity.Property;
import com.susi.susi_suite.entity.Unit;
import org.springframework.stereotype.Component;

@Component
public class UnitConverter {
    public Unit toEntity(UnitRequest request, Property property) {
        if (request == null) {
            return null;
        }

        Unit unit = new Unit();
        unit.setProperty(property);
        unit.setRoomNumber(request.roomNumber());
        unit.setRentAmount(request.rentAmount());
        unit.setSquareFootage(request.squareFootage());
        unit.setStatus(request.status());
        unit.setInteriorAmenities(request.interiorAmenities());
        unit.setImages(request.images());
        return unit;
    }

    public UnitResponse toResponse(Unit unit) {
        if (unit == null) {
            return null;
        }

        return new UnitResponse(
                unit.getId(),
                unit.getProperty().getId(),
                unit.getRoomNumber(),
                unit.getRentAmount(),
                unit.getSquareFootage(),
                unit.getStatus(),
                unit.getInteriorAmenities(),
                unit.getImages()
        );
    }
}
