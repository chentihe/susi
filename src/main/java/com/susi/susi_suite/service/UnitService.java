package com.susi.susi_suite.service;

import com.susi.susi_suite.entity.Unit;
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
    private final UnitRepository unitRepository;

    @Transactional
    public Unit createUnit(Unit unit) {
        return unitRepository.save(unit);
    }

    public Unit getUnitById(Long id) throws NoSuchElementException {
        return unitRepository.findById(id).orElseThrow();
    }
}
