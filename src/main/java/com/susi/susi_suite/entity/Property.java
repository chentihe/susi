package com.susi.susi_suite.entity;

import com.fasterxml.jackson.annotation.JsonManagedReference;
import io.hypersistence.utils.hibernate.type.json.JsonBinaryType;
import jakarta.persistence.*;
import jakarta.persistence.CascadeType;
import jakarta.persistence.Table;
import lombok.Data;
import org.hibernate.annotations.*;
import org.hibernate.type.SqlTypes;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Entity
@Table(name = "properties")
@Data
@EntityListeners(AuditingEntityListener.class)
@SQLDelete(sql = "UPDATE properties SET is_deleted = true WHERE id = ?")
@SQLRestriction("is_deleted = false")
public class Property {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    private String name;

    @Column(columnDefinition = "TEXT")
    private String description;

    private String city;
    private String area;
    private String road;
    private String scope;
    private String address;
    private String zipcode;

    @PrePersist
    @PreUpdate
    public void generateFullAddress() {
        this.address = Stream.of(city, area, road, scope)
                .filter(s -> s != null && !s.isBlank())
                .collect(Collectors.joining(""));
    }

    private Integer totalFloors;
    private Integer floor;
    private Integer basementFloors;

    @Column(name = "landlord_id")
    private Long landlordId;

    @Type(JsonBinaryType.class)
    @JdbcTypeCode(SqlTypes.JSON)
    @Column(name = "public_amenities", columnDefinition = "jsonb")
    private Map<String, Object> publicAmenities;

    @OneToMany(mappedBy = "property", cascade = CascadeType.ALL, fetch = FetchType.LAZY)
    @JsonManagedReference
    private List<Unit> units = new ArrayList<>();

    private String coverImage;

    @Type(JsonBinaryType.class)
    @Column(columnDefinition = "jsonb")
    private List<String> images;

    private boolean isDeleted = false;

    @CreatedDate
    @Column(updatable = false, nullable = false, insertable = false)
    private LocalDateTime createdAt;

    @LastModifiedDate
    @Column(nullable = false, insertable = false)
    private LocalDateTime updatedAt;
}
