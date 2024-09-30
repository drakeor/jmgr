#include "uptime.h"

std::string Uptime::getFormattedUptime() {
    std::ifstream uptimeFile("/proc/uptime");

    if (!uptimeFile.is_open()) {
        return "Error: could not open /proc/uptime.";
    }

    double uptimeSeconds;
    uptimeFile >> uptimeSeconds; // Read the uptime in seconds

    int days = uptimeSeconds / 86400;
    uptimeSeconds = std::fmod(uptimeSeconds, 86400);
    int hours = uptimeSeconds / 3600;
    uptimeSeconds = std::fmod(uptimeSeconds, 3600);
    int minutes = uptimeSeconds / 60;
    int seconds = static_cast<int>(uptimeSeconds) % 60;

    std::string result = "";
    if (days > 0) {
        result += std::to_string(days) + " days, ";
    }
    result += std::to_string(hours) + " hours, " + std::to_string(minutes) + " minutes, " + std::to_string(seconds) + " seconds";

    uptimeFile.close();
    return result;
}
