#include <iostream>
#include <sys/utsname.h>

class KernelInfo {
public:
    // Method to load and print the kernel version
    void loadAndPrintKernelVersion() {
        struct utsname buffer;

        if (uname(&buffer) != 0) {
            std::cerr << "Error: could not get kernel version." << std::endl;
            exit(1);
        }

        std::cout << "Kernel Version: " << buffer.release << std::endl;
    }
};

int main() {
    // Create a KernelInfo object
    KernelInfo kernelInfo;
    
    // Load and print the kernel version
    kernelInfo.loadAndPrintKernelVersion();

    return 0;
}
