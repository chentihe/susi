package com.susi.susi_suite.repository;

import com.susi.susi_suite.entity.Unit;
import com.susi.susi_suite.entity.UnitStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface UnitRepository extends JpaRepository<Unit, Long> {

    List<Unit> findByPropertyIdAndStatus(Long propertyId, UnitStatus status);
}
