#include "process_info.h"
#include <fstream>
#include <unistd.h>
#include <sstream>
#include <iostream>

#include <dirent.h>
#include <algorithm>
#include <iomanip> 

ProcessInfo::ProcessInfo(int processId) : pid(processId), rssMemoryUsage(0), name("") {}

std::vector<ProcessInfo> ProcessInfo::fetchProcesses()
{
    DIR* dir = opendir("/proc");
    struct dirent* entry;
    std::vector<ProcessInfo> processes;

    while ((entry = readdir(dir)) != nullptr) {
        if (entry->d_type == DT_DIR) {
            int pid = atoi(entry->d_name);
            if (pid > 0) {
                ProcessInfo process(pid);
                process.fetchRSSMemoryUsage();
                processes.push_back(process);
            }
        }
    }
    closedir(dir);

    // Sort the processes by memory usage
    std::sort(processes.begin(), processes.end(),
        [](const ProcessInfo& a, const ProcessInfo& b) {
            return a.getRSSMemoryUsage() > b.getRSSMemoryUsage();
    });

    return processes;
}

void ProcessInfo::fetchRSSMemoryUsage() {
    std::ifstream statmFile("/proc/" + std::to_string(pid) + "/statm");
    long unused, rss;
    statmFile >> unused >> rss;
    rssMemoryUsage = rss * sysconf(_SC_PAGESIZE) / 1024; // Convert pages to kB
}

std::string ProcessInfo::getProcessName() const {
    std::ifstream cmdlineFile("/proc/" + std::to_string(pid) + "/comm");
    if (cmdlineFile.is_open()) {
        std::string pname;
        std::getline(cmdlineFile, pname); // pname stores the process name now
        return pname;
    }
    return "[unknown]";
}

long ProcessInfo::getRSSMemoryUsage() const {
    return rssMemoryUsage;
}

std::string ProcessInfo::getRSSMemoryUsageFormat() const {
    return formatMemory(rssMemoryUsage);
}

int ProcessInfo::getPID() const {
    return pid;
}

std::string ProcessInfo::formatMemory(long memoryInKB) const {
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
