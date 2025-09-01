package com.example.workloadorch;

import com.azure.resourcemanager.workloadorchestration.WorkloadOrchestrationManager;
import com.azure.resourcemanager.workloadorchestration.models.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.io.*;
import java.nio.file.*;
import java.util.concurrent.locks.ReentrantLock;

/**
 * Manages version numbers for solution templates and schemas.
 */
public class VersionManager {
    private static final Logger logger = LoggerFactory.getLogger(VersionManager.class);
    private static final String VERSION_FILE = "java/version.txt";
    private static final String DEFAULT_VERSION = "1.1.0";
    private static final ReentrantLock lock = new ReentrantLock();
    
    private final WorkloadOrchestrationManager manager;
    private final String resourceGroup;

    public VersionManager(WorkloadOrchestrationManager manager, String resourceGroup) {
        this.manager = manager;
        this.resourceGroup = resourceGroup;
        initializeVersionFile();
    }

    private void initializeVersionFile() {
        try {
            File file = new File(VERSION_FILE);
            if (!file.exists()) {
                file.getParentFile().mkdirs();
                Files.writeString(file.toPath(), DEFAULT_VERSION);
            }
        } catch (IOException e) {
            logger.error("Error initializing version file: {}", e.getMessage());
        }
    }

    private String incrementVersion(String version) {
        String[] parts = version.split("\\.");
        if (parts.length != 3) {
            return DEFAULT_VERSION;
        }

        int major = Integer.parseInt(parts[0]);
        int minor = Integer.parseInt(parts[1]);
        int patch = Integer.parseInt(parts[2]);

        patch++;
        if (patch > 9) {
            patch = 0;
            minor++;
            if (minor > 9) {
                minor = 0;
                major++;
            }
        }

        return String.format("%d.%d.%d", major, minor, patch);
    }

    private String getNextVersion() {
        lock.lock();
        try {
            String currentVersion = Files.readString(Paths.get(VERSION_FILE)).trim();
            String nextVersion = incrementVersion(currentVersion);
            Files.writeString(Paths.get(VERSION_FILE), nextVersion);
            return nextVersion;
        } catch (IOException e) {
            logger.error("Error managing version file: {}", e.getMessage());
            return DEFAULT_VERSION;
        } finally {
            lock.unlock();
        }
    }

    /**
     * Generates a new version for a solution template.
     * @param templateName The name of the solution template
     * @return A unique version string
     */
    public String getNextTemplateVersion(String templateName) {
        return getNextVersion();
    }

    /**
     * Generates a new version for a schema.
     * @param schemaName The name of the schema
     * @return A unique version string
     */
    public String getNextSchemaVersion(String schemaName) {
        return getNextVersion();
    }
}
