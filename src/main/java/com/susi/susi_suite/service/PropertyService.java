package com.susi.susi_suite.service;

import com.susi.susi_suite.client.AddressServiceClient;
import com.susi.susi_suite.dto.response.PropertyResponse;
import com.susi.susi_suite.entity.Property;
import com.susi.susi_suite.mapper.PropertyConverter;
import com.susi.susi_suite.repository.PropertyRepository;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class PropertyService {
    private final PropertyRepository propertyRepository;
    private final AddressServiceClient addressClient;
    private final PropertyConverter propertyConverter;

    @Transactional
    public PropertyResponse createProperty(Property property) {
        try {
            String zipcode = addressClient.getZipcodes(property.getRoad(), 1);
            property.setZipcode(zipcode);
        } catch (Exception e) {
            log.error("Failed to fetch zipcode from address service", e);
        }

        return propertyConverter.toResponse(propertyRepository.save(property));
    }

    public List<PropertyResponse> getAllProperties() {
        return propertyRepository.findAll()
                .stream().map(p -> propertyConverter.toResponse(p))
                .collect(Collectors.toList());
    }

    public PropertyResponse getPropertyById(Long id) throws NoSuchElementException {
        return propertyConverter.toResponse(propertyRepository.findById(id).orElseThrow());
    }
}
