#ifndef KERNELINFO_H
#define KERNELINFO_H

#include <iostream>
#include <sys/utsname.h>
#include <string>

class KernelInfo {
public:
    // Method to load and return the kernel version as a string
    std::string getKernelVersion();
};

#endif // KERNELINFO_H
