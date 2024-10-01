#include "disk_usage.h"
#include <iostream>
#include <vector>
#include <sys/statvfs.h>
#include <fstream>
#include <sstream>
#include <cstring>
#include <set>
#include <iomanip>
#include <cmath>

// Constructor to initialize the exclude lists
DiskUsage::DiskUsage() : excludeFileSystems{
        "tmpfs", "overlay", "devtmpfs", "proc", "sysfs", "cgroup", "devpts"
    },
    excludeMountPoints{
        "/run", "/var/lib/docker", "/run/user", "/snap", "/sys", "/boot/efi", "/var/snap"
    } {}

// Method to check if a filesystem or mount point should be excluded
bool DiskUsage::shouldExclude(const std::string& fsType, const std::string& mountPoint) {
    // Exclude based on filesystem type
    if (excludeFileSystems.count(fsType) > 0) {
        return true;
    }

    // Exclude based on the mount point path
    for (const std::string& excludePath : excludeMountPoints) {
        if (mountPoint.find(excludePath) == 0) {
            return true;
        }
    }

    return false;
}

// Method to get all relevant partitions from /proc/mounts
std::vector<std::string> DiskUsage::getPartitions() {
    std::vector<std::string> partitions;
    std::ifstream mounts("/proc/mounts");
    std::string line;

    while (std::getline(mounts, line)) {
        std::istringstream iss(line);
        std::string device, mountPoint, fsType, options;
        int dump, pass;
        if (iss >> device >> mountPoint >> fsType >> options >> dump >> pass) {
            if (!shouldExclude(fsType, mountPoint)) {
                partitions.push_back(mountPoint);
            }
        }
    }

    return partitions;
}

// Method to format and print size in appropriate unit
std::string DiskUsage::printSize(unsigned long long bytes) {
    const unsigned long long TB = 1024ull * 1024 * 1024 * 1024;
    const unsigned long long GB = 1024ull * 1024 * 1024;
    const unsigned long long MB = 1024ull * 1024;

    std::ostringstream oss;
    oss << std::fixed << std::setprecision(2);  // Set precision to 2 decimal places

    if (bytes >= TB) {
        oss << (double)bytes / TB << " TB";
    } else if (bytes >= GB) {
        oss << (double)bytes / GB << " GB";
    } else {
        oss << (double)bytes / MB << " MB";
    }

    return oss.str();  // Return the formatted string
}

// Method to print disk usage similar to df command for a given partition
std::string DiskUsage::printDiskUsage(const std::string& partition) {
    struct statvfs stat;
    if (statvfs(partition.c_str(), &stat) != 0) {
        // Silently ignore any partitions that can't be accessed
        return "";
    }

    unsigned long long totalSpace = stat.f_blocks * stat.f_frsize;
    unsigned long long freeSpace = stat.f_bfree * stat.f_frsize;
    unsigned long long usedSpace = totalSpace - freeSpace;

    // Skip partitions with 0 size (invalid/irrelevant)
    if (totalSpace == 0) return "";

    // Calculate percentage used
    long percentUsed = (static_cast<double>(usedSpace) / totalSpace) * 100;

    // Assemble final string
    std::string result = printSize(usedSpace) + " / " 
        + printSize(totalSpace) + " (" + std::to_string(percentUsed) + "%)";
    return result;
}
