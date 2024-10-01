#ifndef PROCESSINFO_H
#define PROCESSINFO_H

#include <string>
#include <vector>

class ProcessInfo {
public:
    ProcessInfo(int pid);

    void fetchRSSMemoryUsage();
    std::string getProcessName() const;
    long getRSSMemoryUsage() const;
    std::string getRSSMemoryUsageFormat() const;
    int getPID() const;

    static std::vector<ProcessInfo> fetchProcesses();
    
private:
    int pid;
    long rssMemoryUsage;
    std::string name;

    std::string formatMemory(long memoryInKB) const;
};

#endif // PROCESSINFO_H
