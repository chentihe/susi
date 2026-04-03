package com.susi.susi_suite.repository;

import com.susi.susi_suite.entity.Property;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface PropertyRepository extends JpaRepository<Property, Long> {

    @EntityGraph(attributePaths = {"units"})
    Optional<Property> findWithUnitsById(Long id);

    List<Property> findByLandlordId(Long landlordId);
}
