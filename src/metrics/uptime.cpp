#include <iostream>
#include <fstream>
#include <string>
#include <cmath>

class Uptime {
public:
    // Method to load and print uptime
    void loadAndPrintUptime() {
        std::ifstream uptimeFile("/proc/uptime");

        if (!uptimeFile.is_open()) {
            std::cerr << "Error: could not open /proc/uptime." << std::endl;
            exit(1);
        }

        double uptimeSeconds;
        uptimeFile >> uptimeSeconds; // Read the uptime in seconds

        int days = uptimeSeconds / 86400;
        uptimeSeconds = std::fmod(uptimeSeconds, 86400);
        int hours = uptimeSeconds / 3600;
        uptimeSeconds = std::fmod(uptimeSeconds, 3600);
        int minutes = uptimeSeconds / 60;
        int seconds = static_cast<int>(uptimeSeconds) % 60;

        std::cout << "System Uptime: ";
        if (days > 0) {
            std::cout << days << " days, ";
        }
        std::cout << hours << " hours, " << minutes << " minutes, " << seconds << " seconds" << std::endl;

        uptimeFile.close();
    }
};

int main() {
    // Create an Uptime object
    Uptime uptime;
    
    // Load and print the system uptime
    uptime.loadAndPrintUptime();

    return 0;
}
