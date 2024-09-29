#include <iostream>
#include <fstream>
#include <string>
#include <sstream>
#include <map>

using namespace std;

struct NetworkData {
    unsigned long long rx_bytes; // Received bytes
    unsigned long long tx_bytes; // Transmitted bytes
};

map<string, NetworkData> get_all_network_data() {
    map<string, NetworkData> network_data;
    std::ifstream file("/proc/net/dev");
    std::string line;

    if (file.is_open()) {
        // Skip the first two lines (header)
        std::getline(file, line);
        std::getline(file, line);

        // Parse each line for the network interface statistics
        while (std::getline(file, line)) {
            std::istringstream ss(line);
            std::string iface;
            unsigned long long rx_bytes, rx_packets, rx_errs, rx_drop, rx_fifo, rx_frame, rx_compressed, rx_multicast;
            unsigned long long tx_bytes, tx_packets, tx_errs, tx_drop, tx_fifo, tx_colls, tx_carrier, tx_compressed;

            // Extract interface name
            ss >> iface;

            // Remove the trailing colon from the interface name
            iface = iface.substr(0, iface.find(':'));

            // Extract the rest of the data
            ss >> rx_bytes >> rx_packets >> rx_errs >> rx_drop >> rx_fifo >> rx_frame >> rx_compressed >> rx_multicast
               >> tx_bytes >> tx_packets >> tx_errs >> tx_drop >> tx_fifo >> tx_colls >> tx_carrier >> tx_compressed;

            // Store data in the map with interface name as key
            NetworkData data = {rx_bytes, tx_bytes};
            network_data[iface] = data;
        }
        file.close();
    }

    return network_data;
}

int main() {
    map<string, NetworkData> network_stats = get_all_network_data();

    cout << "Network traffic statistics:" << endl;
    for (const auto& entry : network_stats) {
        const string& iface = entry.first;
        const NetworkData& data = entry.second;
        cout << "Interface: " << iface << endl;
        cout << "  Received: " << data.rx_bytes / (1024 * 1024) << " MB" << endl;
        cout << "  Transmitted: " << data.tx_bytes / (1024 * 1024) << " MB" << endl;
    }

    return 0;
}
