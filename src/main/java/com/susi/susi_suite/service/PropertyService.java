package com.susi.susi_suite.service;

import com.susi.susi_suite.client.AddressServiceClient;
import com.susi.susi_suite.entity.Property;
import com.susi.susi_suite.repository.PropertyRepository;
import jakarta.transaction.Transactional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.NoSuchElementException;

@Service
@RequiredArgsConstructor
@Slf4j
public class PropertyService {
    private final PropertyRepository propertyRepository;
    private final AddressServiceClient addressClient;

    @Transactional
    public Property createProperty(Property property) {
        try {
            String zipcode = addressClient.getZipcodes(property.getRoad(), 1);
            property.setZipcode(zipcode);
        } catch (Exception e) {
            log.error("Failed to fetch zipcode from address service", e);
        }

        return propertyRepository.save(property);
    }

    public List<Property> getAllProperties() {
        return propertyRepository.findAll();
    }

    public Property getPropertyById(Long id) throws NoSuchElementException {
        return propertyRepository.findById(id).orElseThrow();
    }
}
