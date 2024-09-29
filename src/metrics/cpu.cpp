#include <iostream>
#include <fstream>
#include <string>
#include <thread>
#include <chrono>

class CPUTracker {
public:
    // Struct to store CPU usage data
    struct CPUData {
        unsigned long long int user;
        unsigned long long int nice;
        unsigned long long int system;
        unsigned long long int idle;
        unsigned long long int iowait;
        unsigned long long int irq;
        unsigned long long int softirq;
        unsigned long long int steal;
    };

    // Method to start tracking CPU usage
    void trackCPUUsage() {
        CPUData prevData = getCPUData();
        while (true) {
            std::this_thread::sleep_for(std::chrono::seconds(1)); // Adjust interval as needed
            CPUData currData = getCPUData();
            float usage = calculateCPUUsage(prevData, currData);
            std::cout << "CPU Usage: " << usage << "%" << std::endl;
            prevData = currData; // Update previous data
        }
    }

private:
    // Method to parse CPU data from /proc/stat
    CPUData parseCPUData(const std::string& cpuLine) const {
        CPUData data;
        sscanf(cpuLine.c_str(), "cpu  %llu %llu %llu %llu %llu %llu %llu %llu",
               &data.user, &data.nice, &data.system, &data.idle,
               &data.iowait, &data.irq, &data.softirq, &data.steal);
        return data;
    }

    // Method to get CPU data from /proc/stat
    CPUData getCPUData() const {
        std::ifstream file("/proc/stat");
        std::string line;
        if (file.is_open()) {
            std::getline(file, line); // Read the first line for overall CPU stats
            file.close();
        }
        return parseCPUData(line);
    }

    // Method to calculate CPU usage percentage
    float calculateCPUUsage(const CPUData& prev, const CPUData& curr) const {
        unsigned long long int prevIdle = prev.idle + prev.iowait;
        unsigned long long int currIdle = curr.idle + curr.iowait;

        unsigned long long int prevNonIdle = prev.user + prev.nice + prev.system +
                                             prev.irq + prev.softirq + prev.steal;
        unsigned long long int currNonIdle = curr.user + curr.nice + curr.system +
                                             curr.irq + curr.softirq + curr.steal;

        unsigned long long int prevTotal = prevIdle + prevNonIdle;
        unsigned long long int currTotal = currIdle + currNonIdle;

        unsigned long long int totalDiff = currTotal - prevTotal;
        unsigned long long int idleDiff = currIdle - prevIdle;

        return (totalDiff - idleDiff) * 100.0f / totalDiff;
    }
};

int main() {
    CPUTracker cpuTracker;
    cpuTracker.trackCPUUsage();
    return 0;
}
