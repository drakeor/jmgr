#include <iostream>
#include <fstream>
#include <string>
#include <cstdio>
#include <cstdlib>

class CPUInfo {
public:
    // Method to load and print CPU details (model, number of cores, clock speed, and max clock speed)
    void loadAndPrintCPUInfo() {
        std::ifstream cpuInfoFile("/proc/cpuinfo");

        if (!cpuInfoFile.is_open()) {
            std::cerr << "Error: could not open /proc/cpuinfo." << std::endl;
            exit(1);
        }

        std::string line;
        std::string model;
        int numCores = 0;
        double clockSpeed = 0.0;
        double maxClockSpeed = getMaxClockSpeed();

        while (std::getline(cpuInfoFile, line)) {
            if (line.find("model name") == 0) {
                if (model.empty()) {
                    model = line.substr(line.find(":") + 2);  // Extract CPU model after ": "
                }
            } else if (line.find("cpu cores") == 0) {
                numCores = std::stoi(line.substr(line.find(":") + 2));  // Extract the number of cores
            } else if (line.find("cpu MHz") == 0) {
                clockSpeed = std::stod(line.substr(line.find(":") + 2));  // Extract the current clock speed in MHz
            }
        }

        std::cout << "CPU Model: " << model << std::endl;
        std::cout << "Number of Cores: " << numCores << std::endl;
        std::cout << "Current Clock Speed: " << clockSpeed << " MHz" << std::endl;
        std::cout << "Maximum Clock Speed (Stated): " << maxClockSpeed << " MHz" << std::endl;

        cpuInfoFile.close();
    }

private:
    // Method to get the maximum clock speed (stated speed)
    double getMaxClockSpeed() {
        // Try reading from /sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq
        double maxClockSpeed = readMaxClockSpeedFromSys();
        if (maxClockSpeed > 0) {
            return maxClockSpeed;
        }

        // Fallback to lscpu command if /sys/... doesn't work
        maxClockSpeed = readMaxClockSpeedFromLscpu();
        if (maxClockSpeed > 0) {
            return maxClockSpeed;
        }

        std::cerr << "Warning: Could not determine maximum clock speed." << std::endl;
        return 0.0;
    }

    // Method to read max clock speed from /sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq
    double readMaxClockSpeedFromSys() {
        std::ifstream maxFreqFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq");
        double maxClockSpeed = 0.0;

        if (maxFreqFile.is_open()) {
            maxFreqFile >> maxClockSpeed;
            maxFreqFile.close();
            return maxClockSpeed / 1000; // Return in MHz
        }
        return 0.0;  // Return 0.0 if the file is not accessible
    }

    // Method to read max clock speed from lscpu
    double readMaxClockSpeedFromLscpu() {
        FILE* pipe = popen("lscpu | grep 'MHz' | grep 'max'", "r");
        if (!pipe) {
            return 0.0;
        }

        char buffer[128];
        std::string result = "";

        while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
            result += buffer;
        }

        pclose(pipe);

        // Extract the MHz value from the output
        size_t pos = result.find_last_of(" ");
        if (pos != std::string::npos) {
            return std::stod(result.substr(pos + 1));
        }

        return 0.0;  // Return 0.0 if parsing failed
    }
};

int main() {
    // Create a CPUInfo object
    CPUInfo cpuInfo;

    // Load and print the CPU information
    cpuInfo.loadAndPrintCPUInfo();

    return 0;
}
