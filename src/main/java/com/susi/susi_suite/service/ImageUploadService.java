package com.susi.susi_suite.service;

import lombok.extern.slf4j.Slf4j;
import net.coobird.thumbnailator.Thumbnails;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.UUID;

@Service
@Slf4j
public class ImageUploadService {

    @Value("${upload.path:./uploads/images}")
    private String uploadPath;

    public String uploadImage(MultipartFile file) throws IOException {
        // Check if file exists
        if (file.isEmpty()) {
            throw new IllegalArgumentException("File is empty");
        }

        // Check content type
        String contentType = file.getContentType();
        if (contentType == null || !contentType.startsWith("image/")) {
            throw new IllegalArgumentException("Only images are allowed");
        }

        // Check folder exists
        Path root = Paths.get(uploadPath);
        if (!Files.exists(root)) {
            Files.createDirectories(root);
        }

        // UUID + extension to avoid conflict
        String originalFilename = file.getOriginalFilename();
        String extension = originalFilename.substring(originalFilename.lastIndexOf("."));
        String newFilename = UUID.randomUUID().toString() + extension;
        File destinationFile = root.resolve(newFilename).toFile();

        Thumbnails.of(file.getInputStream())
                        .size(1024, 1024)
                                .outputQuality(0.85)
                                        .toFile(destinationFile);

        log.info("Image uploaded successfully: {}", newFilename);
        return "/image/" + newFilename;
    }
}
