#include <iostream>
#include <fstream>
#include <string>
#include <cmath>

class Uptime {
public:
    // Method to get the system uptime as a formatted string
    std::string getFormattedUptime() {
        std::ifstream uptimeFile("/proc/uptime");

        if (!uptimeFile.is_open()) {
            throw std::runtime_error("Error: could not open /proc/uptime.");
        }

        double uptimeSeconds;
        uptimeFile >> uptimeSeconds; // Read the uptime in seconds

        int days = uptimeSeconds / 86400;
        uptimeSeconds = std::fmod(uptimeSeconds, 86400);
        int hours = uptimeSeconds / 3600;
        uptimeSeconds = std::fmod(uptimeSeconds, 3600);
        int minutes = uptimeSeconds / 60;
        int seconds = static_cast<int>(uptimeSeconds) % 60;

        uptimeFile.close();

        // Format the uptime as a string
        std::string formattedUptime = "System Uptime: ";
        if (days > 0) {
            formattedUptime += std::to_string(days) + " days, ";
        }
        formattedUptime += std::to_string(hours) + " hours, " +
                           std::to_string(minutes) + " minutes, " +
                           std::to_string(seconds) + " seconds";

        return formattedUptime;
    }

    // Method to print the system uptime
    void printUptime() {
        try {
            std::string uptime = getFormattedUptime();
            std::cout << uptime << std::endl;
        } catch (const std::runtime_error& e) {
            std::cerr << e.what() << std::endl;
        }
    }
};

int main() {
    // Create an Uptime object
    Uptime uptime;

    // Print the system uptime
    uptime.printUptime();

    return 0;
}
