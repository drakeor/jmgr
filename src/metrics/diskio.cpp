#include <iostream>
#include <fstream>
#include <string>
#include <sstream>
#include <unistd.h>
#include <vector>

class DiskStats {
public:
    // Struct to store disk I/O statistics
    struct DiskIO {
        std::string deviceName;
        unsigned long readsCompleted;
        unsigned long readsMerged;
        unsigned long sectorsRead;
        unsigned long readTime;
        unsigned long writesCompleted;
        unsigned long writesMerged;
        unsigned long sectorsWritten;
        unsigned long writeTime;
        unsigned long ioInProgress;
        unsigned long ioTime;
        unsigned long weightedIOTime;
    };

    // Method to load disk I/O stats
    std::vector<DiskIO> loadDiskIOStats() {
        std::ifstream diskStatsFile("/proc/diskstats");
        std::vector<DiskIO> diskStats;
        if (!diskStatsFile.is_open()) {
            std::cerr << "Error: could not open /proc/diskstats." << std::endl;
            return diskStats;
        }

        std::string line;
        while (std::getline(diskStatsFile, line)) {
            std::istringstream iss(line);
            DiskIO diskIO;
            int majorNumber, minorNumber;

            // Extract the disk statistics from the line
            iss >> majorNumber >> minorNumber >> diskIO.deviceName 
                >> diskIO.readsCompleted >> diskIO.readsMerged 
                >> diskIO.sectorsRead >> diskIO.readTime 
                >> diskIO.writesCompleted >> diskIO.writesMerged 
                >> diskIO.sectorsWritten >> diskIO.writeTime 
                >> diskIO.ioInProgress >> diskIO.ioTime >> diskIO.weightedIOTime;

            diskStats.push_back(diskIO);
        }

        diskStatsFile.close();
        return diskStats;
    }

    // Method to print disk I/O statistics
    void printDiskIOStats(const std::vector<DiskIO>& diskStats) {
        std::cout << "Disk I/O Statistics:" << std::endl;
        for (const auto& diskIO : diskStats) {
            // Skip loopback devices or irrelevant devices (you can modify this logic to suit your needs)
            if (diskIO.deviceName.find("loop") == std::string::npos) {
                std::cout << "Device: " << diskIO.deviceName << std::endl;
                std::cout << "Reads Completed: " << diskIO.readsCompleted << std::endl;
                std::cout << "Writes Completed: " << diskIO.writesCompleted << std::endl;
                std::cout << "Sectors Read: " << diskIO.sectorsRead << std::endl;
                std::cout << "Sectors Written: " << diskIO.sectorsWritten << std::endl;
                std::cout << "Time Spent Reading (ms): " << diskIO.readTime << std::endl;
                std::cout << "Time Spent Writing (ms): " << diskIO.writeTime << std::endl;
                std::cout << "IO in Progress: " << diskIO.ioInProgress << std::endl;
                std::cout << "Total IO Time (ms): " << diskIO.ioTime << std::endl;
                std::cout << "Weighted IO Time (ms): " << diskIO.weightedIOTime << std::endl;
                std::cout << "---------------------------------------" << std::endl;
            }
        }
    }
};

int main() {
    DiskStats diskStatsMonitor;

    // Load the disk I/O stats
    std::vector<DiskStats::DiskIO> diskStats = diskStatsMonitor.loadDiskIOStats();

    // Print the disk I/O stats
    diskStatsMonitor.printDiskIOStats(diskStats);

    return 0;
}
