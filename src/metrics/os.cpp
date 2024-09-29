#include <iostream>
#include <fstream>
#include <string>

class OSInfo {
public:
    // Method to load and print OS information
    void loadAndPrintOSInfo() {
        std::ifstream osFile("/etc/os-release");

        if (!osFile.is_open()) {
            std::cerr << "Error: could not open /etc/os-release." << std::endl;
            exit(1);
        }

        std::string line;
        while (std::getline(osFile, line)) {
            if (line.find("PRETTY_NAME=") == 0) {
                std::string osName = line.substr(13, line.length() - 14);  // Remove the surrounding quotes
                std::cout << "Operating System: " << osName << std::endl;
                break;
            }
        }

        osFile.close();
    }
};

int main() {
    // Create an OSInfo object
    OSInfo osInfo;
    
    // Load and print the OS information
    osInfo.loadAndPrintOSInfo();

    return 0;
}
