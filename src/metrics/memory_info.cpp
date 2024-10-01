#include "memory_info.h"
#include <fstream>
#include <sstream>
#include <iomanip>

MemoryInfo::MemoryInfo() : totalMem(0), freeMem(0), buffers(0), cached(0), inUse(0) {}

void MemoryInfo::fetchMemoryInfo() {
    std::ifstream meminfo("/proc/meminfo");
    std::string line;

    while (std::getline(meminfo, line)) {
        std::istringstream iss(line);
        std::string key;
        long value;
        iss >> key >> value;

        if (key == "MemTotal:") totalMem = value;
        else if (key == "MemFree:") freeMem = value;
        else if (key == "Buffers:") buffers = value;
        else if (key == "Cached:") cached = value;
    }

    inUse = totalMem - (freeMem + buffers + cached);
}

std::string MemoryInfo::formatMemory(long memoryInKB) const {
    if (memoryInKB >= 1024 * 1024) {
        double memoryInGB = static_cast<double>(memoryInKB) / (1024 * 1024);
        std::ostringstream oss;
        oss << std::fixed << std::setprecision(2) << memoryInGB << " GB";
        return oss.str();
    } else if (memoryInKB >= 1024) {
        double memoryInMB = static_cast<double>(memoryInKB) / 1024;
        std::ostringstream oss;
        oss << std::fixed << std::setprecision(2) << memoryInMB << " MB";
        return oss.str();
    } else {
        return std::to_string(memoryInKB) + " kB";
    }
}

std::string MemoryInfo::getFormattedTotalMemory() const
{
    std::string result = getMemoryInUse() 
        + " / " + getTotalMemory() 
        + " (" + std::to_string(getMemoryUsePercent()) + "%)";
    return result;
}

long MemoryInfo::getMemoryUsePercent() const {
    return (inUse * 100) / totalMem;
}

std::string MemoryInfo::getTotalMemory() const {
    return formatMemory(totalMem);
}

std::string MemoryInfo::getCachedMemory() const {
    return formatMemory(cached);
}

std::string MemoryInfo::getBufferMemory() const {
    return formatMemory(buffers);
}

std::string MemoryInfo::getMemoryInUse() const {
    return formatMemory(inUse);
}
