package com.susi.susi_suite.service;

import com.susi.susi_suite.dto.request.UnitRequest;
import com.susi.susi_suite.dto.response.UnitResponse;
import com.susi.susi_suite.entity.Property;
import com.susi.susi_suite.entity.Unit;
import com.susi.susi_suite.mapper.UnitConverter;
import com.susi.susi_suite.repository.PropertyRepository;
import com.susi.susi_suite.repository.UnitRepository;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.NoSuchElementException;

@Service
@RequiredArgsConstructor
@Slf4j
public class UnitService {
    private final PropertyRepository propertyRepository;
    private final UnitRepository unitRepository;
    private final UnitConverter unitConverter;

    @Transactional
    public UnitResponse createUnit(UnitRequest request) {
        Property property = propertyRepository.findById(request.propertyId())
                .orElseThrow(() -> new RuntimeException("Property not found"));

        Unit entity = unitConverter.toEntity(request, property);

        Unit unit = unitRepository.save(entity);

        return unitConverter.toResponse(unit);
    }

    public UnitResponse getUnitById(Long id) throws NoSuchElementException {
        return unitConverter.toResponse(unitRepository.findById(id).orElseThrow());
    }
}
