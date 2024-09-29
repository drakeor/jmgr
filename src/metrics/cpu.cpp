#include <iostream>
#include <fstream>
#include <string>
#include <thread>
#include <chrono>

using namespace std;

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

CPUData parse_cpu_data(const std::string& cpu_line) {
    CPUData data;
    sscanf(cpu_line.c_str(), "cpu  %llu %llu %llu %llu %llu %llu %llu %llu",
           &data.user, &data.nice, &data.system, &data.idle,
           &data.iowait, &data.irq, &data.softirq, &data.steal);
    return data;
}

CPUData get_cpu_data() {
    std::ifstream file("/proc/stat");
    std::string line;
    if (file.is_open()) {
        std::getline(file, line); // Read the first line for overall CPU stats
        file.close();
    }
    return parse_cpu_data(line);
}

float calculate_cpu_usage(const CPUData& old_data, const CPUData& new_data) {
    unsigned long long int old_idle = old_data.idle + old_data.iowait;
    unsigned long long int new_idle = new_data.idle + new_data.iowait;

    unsigned long long int old_total = old_data.user + old_data.nice + old_data.system + old_data.idle +
                                       old_data.iowait + old_data.irq + old_data.softirq + old_data.steal;
    unsigned long long int new_total = new_data.user + new_data.nice + new_data.system + new_data.idle +
                                       new_data.iowait + new_data.irq + new_data.softirq + new_data.steal;

    unsigned long long int total_diff = new_total - old_total;
    unsigned long long int idle_diff = new_idle - old_idle;

    return (1.0 - (float)idle_diff / total_diff) * 100.0;
}

int main() {
    CPUData old_data = get_cpu_data();
    this_thread::sleep_for(chrono::milliseconds(1000)); // Sleep for 1 second
    CPUData new_data = get_cpu_data();

    float cpu_usage = calculate_cpu_usage(old_data, new_data);
    cout << "CPU Usage: " << cpu_usage << "%" << endl;

    return 0;
}
