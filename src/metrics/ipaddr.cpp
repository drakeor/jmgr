#include <iostream>
#include <sys/types.h>
#include <ifaddrs.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <cstring>

class NetworkInfo {
public:
    // Method to load and print the local non-loopback IP addresses
    void loadAndPrintLocalIPAddress() {
        struct ifaddrs *interfaces = nullptr;
        struct ifaddrs *ifa = nullptr;
        void *tmpAddrPtr = nullptr;

        // Get a linked list of network interfaces
        if (getifaddrs(&interfaces) == -1) {
            std::cerr << "Error: unable to get network interfaces." << std::endl;
            exit(1);
        }

        std::cout << "Local Non-Loopback IP Addresses: " << std::endl;

        // Loop through the list of interfaces
        for (ifa = interfaces; ifa != nullptr; ifa = ifa->ifa_next) {
            if (!ifa->ifa_addr) {
                continue;
            }

            // Check if the address is IPv4
            if (ifa->ifa_addr->sa_family == AF_INET) {
                tmpAddrPtr = &((struct sockaddr_in *)ifa->ifa_addr)->sin_addr;
                char addressBuffer[INET_ADDRSTRLEN];
                inet_ntop(AF_INET, tmpAddrPtr, addressBuffer, INET_ADDRSTRLEN);

                // Skip loopback address (127.0.0.1)
                if (strcmp(addressBuffer, "127.0.0.1") == 0) {
                    continue;
                }

                // Print the interface name and the IP address
                std::cout << ifa->ifa_name << ": " << addressBuffer << std::endl;
            }
        }

        // Free the memory allocated by getifaddrs
        freeifaddrs(interfaces);
    }
};

int main() {
    // Create a NetworkInfo object
    NetworkInfo networkInfo;

    // Load and print the local non-loopback IP address
    networkInfo.loadAndPrintLocalIPAddress();

    return 0;
}
