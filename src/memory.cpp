#include <iostream>
#include <fstream>
#include <string>

using namespace std;

struct MemoryData {
    unsigned long long total;
    unsigned long long free;
    unsigned long long available;
    unsigned long long buffers;
    unsigned long long cached;
};

MemoryData get_memory_data() {
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

float calculate_memory_usage(const MemoryData& data) {
    unsigned long long used_memory = data.total - data.free - data.buffers - data.cached;
    return (used_memory / (float)data.total) * 100.0;
}

int main() {
    MemoryData mem_data = get_memory_data();
    float mem_usage = calculate_memory_usage(mem_data);

    cout << "Memory Usage: " << mem_usage << "%" << endl;

    return 0;
}
