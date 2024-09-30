#ifndef CPUINFO_H
#define CPUINFO_H

#include <iostream>
#include <fstream>
#include <string>

class CPUInfo {
public:
    // Method to load and return CPU details as a string
    std::string loadAndPrintCPUInfo();

    // Method to load and return only the CPU model and core count as a string
    std::string getModelAndCores();

private:
    // Method to get the maximum clock speed (stated speed)
    double getMaxClockSpeed();

    // Method to read max clock speed from /sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq
    double readMaxClockSpeedFromSys();

    // Method to read max clock speed from lscpu
    double readMaxClockSpeedFromLscpu();
};

#endif // CPUINFO_H
