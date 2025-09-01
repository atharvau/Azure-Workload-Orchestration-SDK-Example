#!/bin/bash

# Build the project
echo "Building project..."
mvn clean package

# Run the application
echo "Running WorkloadOrchestrationDemo..."
mvn exec:java -Dexec.mainClass="com.example.workloadorch.WorkloadOrchestrationDemo"