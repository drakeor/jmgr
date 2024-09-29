#include <iostream>
#include <fstream>
#include <string>
#include <sys/statvfs.h>
#include <vector>
#include <sstream>

class DiskMonitor {
public:
    // Struct to store disk usage information
    struct DiskUsage {
        std::string mount_point;
        unsigned long long total_space;
        unsigned long long used_space;
        unsigned long long free_space;
    };

    // Method to start disk usage monitoring
    void monitorDiskUsage() {
        std::vector<std::string> filesystems = getMountedFilesystems();
        for (const auto& fs : filesystems) {
            DiskUsage usage = getDiskUsage(fs);
            printDiskUsage(usage);
        }
    }

private:
    // Method to retrieve a list of mounted filesystems
    std::vector<std::string> getMountedFilesystems() const {
        std::vector<std::string> filesystems;
        std::ifstream file("/proc/mounts");
        std::string line;

        if (file.is_open()) {
            while (getline(file, line)) {
                std::istringstream iss(line);
                std::string device, mount_point, filesystem_type;
                iss >> device >> mount_point >> filesystem_type;

                // Ignore pseudo filesystems like tmpfs, proc, sysfs, etc.
                if (filesystem_type != "tmpfs" && filesystem_type != "proc" &&
                    filesystem_type != "sysfs" && filesystem_type != "devtmpfs" &&
                    filesystem_type != "devpts" && filesystem_type != "overlay") {
                    filesystems.push_back(mount_point);
                }
            }
            file.close();
        }

        return filesystems;
    }

    // Method to get disk usage stats for a given mount point
    DiskUsage getDiskUsage(const std::string& mount_point) const {
        struct statvfs vfs;

        if (statvfs(mount_point.c_str(), &vfs) != 0) {
            std::cerr << "Error: unable to get stats for mount point: " << mount_point << std::endl;
            return DiskUsage{mount_point, 0, 0, 0};
        }

        unsigned long long total = vfs.f_blocks * vfs.f_frsize;
        unsigned long long free = vfs.f_bfree * vfs.f_frsize;
        unsigned long long used = total - free;

        return DiskUsage{mount_point, total, used, free};
    }

    // Method to print disk usage
    void printDiskUsage(const DiskUsage& usage) const {
        std::cout << "Mount Point: " << usage.mount_point << std::endl;
        std::cout << "Total Space: " << usage.total_space / (1024 * 1024) << " MB" << std::endl;
        std::cout << "Used Space: " << usage.used_space / (1024 * 1024) << " MB" << std::endl;
        std::cout << "Free Space: " << usage.free_space / (1024 * 1024) << " MB" << std::endl;
        std::cout << "-----------------------------" << std::endl;
    }
};

int main() {
    DiskMonitor monitor;

    // Start monitoring disk usage
    monitor.monitorDiskUsage();

    return 0;
}
