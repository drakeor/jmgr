#include <iostream>
#include <fstream>
#include <string>
#include <sys/statvfs.h>
#include <vector>
#include <sstream>

using namespace std;

struct DiskUsage {
    string mount_point;
    unsigned long long total_space;
    unsigned long long used_space;
    unsigned long long free_space;
};

vector<string> get_mounted_filesystems() {
    vector<string> filesystems;
    ifstream file("/proc/mounts");
    string line;

    if (file.is_open()) {
        while (getline(file, line)) {
            istringstream iss(line);
            string device, mount_point, filesystem_type;
            iss >> device >> mount_point >> filesystem_type;

            // Ignore certain pseudo filesystems like tmpfs, proc, sysfs, etc.
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

DiskUsage get_disk_usage(const string& path) {
    struct statvfs stat;
    DiskUsage usage = {path, 0, 0, 0};

    if (statvfs(path.c_str(), &stat) == 0) {
        unsigned long long total_space = stat.f_blocks * stat.f_frsize;
        unsigned long long free_space = stat.f_bfree * stat.f_frsize;
        unsigned long long available_space = stat.f_bavail * stat.f_frsize;
        unsigned long long used_space = total_space - free_space;

        usage.total_space = total_space;
        usage.free_space = free_space;
        usage.used_space = used_space;
    }

    return usage;
}

int main() {
    vector<string> filesystems = get_mounted_filesystems();

    cout << "Disk Usage across all mounted filesystems:" << endl;
    for (const auto& fs : filesystems) {
        DiskUsage usage = get_disk_usage(fs);
        if (usage.total_space > 0) {
            cout << "Mount Point: " << usage.mount_point << endl;
            cout << "  Total Space: " << usage.total_space / (1024 * 1024 * 1024) << " GB" << endl;
            cout << "  Used Space: " << usage.used_space / (1024 * 1024 * 1024) << " GB" << endl;
            cout << "  Free Space: " << usage.free_space / (1024 * 1024 * 1024) << " GB" << endl;
        }
    }

    return 0;
}
