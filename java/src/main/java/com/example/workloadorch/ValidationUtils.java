package com.example.workloadorch;

import java.util.regex.Pattern;

/**
 * Utility class for validation operations.
 */
public final class ValidationUtils {
    private static final Pattern VERSION_PATTERN = Pattern.compile("^\\d+\\.\\d+\\.\\d+$");
    
    private ValidationUtils() {
        // Private constructor to prevent instantiation
    }
    
    /**
     * Validates a semantic version string.
     * @param version Version string to validate
     * @return The validated version string
     * @throws IllegalArgumentException if version format is invalid
     */
    public static String validateVersion(String version) {
        if (version == null || version.trim().isEmpty()) {
            throw new IllegalArgumentException("Version cannot be null or empty");
        }
        
        if (!VERSION_PATTERN.matcher(version).matches()) {
            throw new IllegalArgumentException(
                "Invalid version format. Must be in format: major.minor.patch");
        }
        
        return version;
    }

    /**
     * Validates a configuration string is not empty.
     * @param config Configuration string to validate
     * @return The validated configuration string
     * @throws IllegalArgumentException if config is null or empty
     */
    public static String validateConfiguration(String config) {
        if (config == null || config.trim().isEmpty()) {
            throw new IllegalArgumentException("Configuration cannot be null or empty");
        }
        return config;
    }

    /**
     * Validates that a string value is not null or empty.
     * @param value String to validate
     * @param fieldName Name of the field for error message
     * @return The validated string
     * @throws IllegalArgumentException if value is null or empty
     */
    public static String validateRequired(String value, String fieldName) {
        if (value == null || value.trim().isEmpty()) {
            throw new IllegalArgumentException(fieldName + " cannot be null or empty");
        }
        return value.trim();
    }
}
