#include "cpu_info.h"

std::string CPUInfo::loadAndPrintCPUInfo() {
    std::ifstream cpuInfoFile("/proc/cpuinfo");

    if (!cpuInfoFile.is_open()) {
        return "Error: could not open /proc/cpuinfo.";
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

    cpuInfoFile.close();

    std::string result = "CPU Model: " + model + "\n";
    result += "Number of Cores: " + std::to_string(numCores) + "\n";
    result += "Current Clock Speed: " + std::to_string(clockSpeed) + " MHz\n";
    result += "Maximum Clock Speed (Stated): " + std::to_string(maxClockSpeed) + " MHz\n";

    return result;
}

std::string CPUInfo::getModelAndCores() {
    std::ifstream cpuInfoFile("/proc/cpuinfo");

    if (!cpuInfoFile.is_open()) {
        return "Error: could not open /proc/cpuinfo.";
    }

    std::string line;
    std::string model;
    int numCores = 0;

    while (std::getline(cpuInfoFile, line)) {
        if (line.find("model name") == 0) {
            if (model.empty()) {
                model = line.substr(line.find(":") + 2);  // Extract CPU model after ": "
            }
        } else if (line.find("cpu cores") == 0) {
            numCores = std::stoi(line.substr(line.find(":") + 2));  // Extract the number of cores
        }
    }

    cpuInfoFile.close();

    return model + " (" + std::to_string(numCores) + " cores)";
}

double CPUInfo::getMaxClockSpeed() {
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

double CPUInfo::readMaxClockSpeedFromSys() {
    std::ifstream maxFreqFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq");
    double maxClockSpeed = 0.0;

    if (maxFreqFile.is_open()) {
        maxFreqFile >> maxClockSpeed;
        maxFreqFile.close();
        return maxClockSpeed / 1000; // Return in MHz
    }
    return 0.0;  // Return 0.0 if the file is not accessible
}

double CPUInfo::readMaxClockSpeedFromLscpu() {
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
