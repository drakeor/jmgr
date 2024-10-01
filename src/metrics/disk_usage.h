#ifndef DISK_USAGE_H
#define DISK_USAGE_H

#include <vector>
#include <string>
#include <set>

class DiskUsage {
public:
    // Constructor
    DiskUsage();

    // Method to print disk usage similar to df command for a given partition
    std::string printDiskUsage(const std::string& partition);

    // Method to get all relevant partitions from /proc/mounts
    std::vector<std::string> getPartitions();

private:
    // List of filesystems and mount points to exclude
    const std::set<std::string> excludeFileSystems;
    const std::set<std::string> excludeMountPoints;

    // Method to check if a filesystem or mount point should be excluded
    bool shouldExclude(const std::string& fsType, const std::string& mountPoint);

    // Method to format and print size in appropriate unit
    std::string  printSize(unsigned long long bytes);
};

#endif // DISK_USAGE_H
