#include <iostream>
#include <fstream>
#include <string>

class MemoryMonitor {
public:
    // Struct to store memory data
    struct MemoryData {
        unsigned long long total;
        unsigned long long free;
        unsigned long long available;
        unsigned long long buffers;
        unsigned long long cached;
    };

    // Method to get memory data from /proc/meminfo
    MemoryData getMemoryData() const {
        MemoryData data = {0, 0, 0, 0, 0};
        std::ifstream file("/proc/meminfo");
        std::string line;

        if (file.is_open()) {
            while (std::getline(file, line)) {
                if (line.find("MemTotal:") == 0) {
                    sscanf(line.c_str(), "MemTotal: %llu kB", &data.total);
                } else if (line.find("MemFree:") == 0) {
                    sscanf(line.c_str(), "MemFree: %llu kB", &data.free);
                } else if (line.find("MemAvailable:") == 0) {
                    sscanf(line.c_str(), "MemAvailable: %llu kB", &data.available);
                } else if (line.find("Buffers:") == 0) {
                    sscanf(line.c_str(), "Buffers: %llu kB", &data.buffers);
                } else if (line.find("Cached:") == 0) {
                    sscanf(line.c_str(), "Cached: %llu kB", &data.cached);
                }
            }
            file.close();
        }

        return data;
    }

    // Method to calculate memory usage percentage
    float calculateMemoryUsage(const MemoryData& data) const {
