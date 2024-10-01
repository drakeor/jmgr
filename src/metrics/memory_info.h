#ifndef MEMORYINFO_H
#define MEMORYINFO_H

#include <string>

class MemoryInfo {
public:
    MemoryInfo();

    void fetchMemoryInfo();
    std::string getTotalMemory() const;
    std::string getCachedMemory() const;
    std::string getBufferMemory() const;
    std::string getMemoryInUse() const;

    long getMemoryUsePercent() const;

    std::string getFormattedTotalMemory() const;

private:
    long totalMem;
    long freeMem;
    long buffers;
    long cached;
    long inUse;

    std::string formatMemory(long memoryInKB) const;
};

#endif // MEMORYINFO_H
