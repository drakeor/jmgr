#include "os_info.h"

std::string OSInfo::getOSInfo() {
    std::ifstream osFile("/etc/os-release");

    if (!osFile.is_open()) {
        return "Error: could not open /etc/os-release.";
    }

    std::string line;
    std::string osName;
    while (std::getline(osFile, line)) {
        if (line.find("PRETTY_NAME=") == 0) {
            osName = line.substr(13, line.length() - 14);  // Remove the surrounding quotes
            break;
        }
    }

    osFile.close();
    return osName.empty() ? "Error: could not find OS name." : osName;
}
