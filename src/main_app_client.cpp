#include <iostream>
#include <zmq.hpp>
#include <string>
#include <vector>
#include <chrono>
#include <thread>
#include <optional>

const int TIMEOUT_MS = 3000;  // 3-second timeout to detect failures

// Function to connect to multiple servers
void connect_to_servers(zmq::context_t& context, std::vector<std::string>& server_addresses, std::vector<zmq::socket_t>& subscribers) {
    for (const auto& address : server_addresses) {
        zmq::socket_t subscriber(context, zmq::socket_type::sub);
        subscriber.setsockopt(ZMQ_SUBSCRIBE, "", 0); // Subscribe to all messages
        try {
            subscriber.connect(address);
            subscribers.push_back(std::move(subscriber));
            std::cout << "Connected to server: " << address << std::endl;
        } catch (const zmq::error_t& e) {
            std::cerr << "Failed to connect to server: " << address << " (" << e.what() << ")" << std::endl;
        }
    }
}

int main() {
    zmq::context_t context(1); // 1 I/O thread

    // List of server addresses
    std::vector<std::string> server_addresses = {
        "tcp://localhost:5555",
        "tcp://localhost:5556",
        "tcp://localhost:5557"  // Add more server addresses as needed
    };

    std::vector<zmq::socket_t> subscribers;

    // Connect to multiple servers
    connect_to_servers(context, server_addresses, subscribers);

    // Polling setup for asynchronous receiving
    std::vector<zmq::pollitem_t> poll_items;
    for (auto& subscriber : subscribers) {
        zmq::pollitem_t item = { static_cast<void*>(subscriber), 0, ZMQ_POLLIN, 0 };
        poll_items.push_back(item);
    }

    std::vector<int> no_response_counters(server_addresses.size(), 0);  // Track time without response

    // Async polling loop
    while (true) {
        try {
            int poll_result = zmq::poll(poll_items.data(), poll_items.size(), TIMEOUT_MS);  // Poll with timeout

            if (poll_result == -1) {
                // Error occurred during poll
                throw zmq::error_t();
            }

            for (size_t i = 0; i < poll_items.size(); ++i) {
                if (poll_items[i].revents & ZMQ_POLLIN) {
                    zmq::message_t message;
                    try {
                        auto result = subscribers[i].recv(message, zmq::recv_flags::none);
                        if (result.has_value()) {  // Check if we successfully received a message
                            std::string received_data(static_cast<char*>(message.data()), message.size());
                            std::cout << "Received from server " << server_addresses[i] << ": " << received_data << std::endl;

                            // Reset the counter for this server since we received data
                            no_response_counters[i] = 0;
                        } else {
                            std::cerr << "Error: recv() did not return a valid result, server " << server_addresses[i] << " may have a problem." << std::endl;
                        }
                    } catch (const zmq::error_t& e) {
                        std::cerr << "Error receiving from server " << server_addresses[i] << ": " << e.what() << std::endl;
                    }
                } else {
                    // No data received, increment the failure counter
                    no_response_counters[i] += TIMEOUT_MS;
                    if (no_response_counters[i] >= TIMEOUT_MS * 3) {  // After 3 failed polls
                        std::cerr << "Warning: No response from server " << server_addresses[i] << " for " 
                                  << no_response_counters[i] / 1000 << " seconds." << std::endl;
                    }
                }
            }
        } catch (const zmq::error_t& e) {
            std::cerr << "Error during polling: " << e.what() << std::endl;
        }

        // Optionally, add a small sleep to prevent busy looping
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }

    return 0;
}
