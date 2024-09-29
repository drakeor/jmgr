#include <iostream>
#include <zmq.hpp>
#include <string>
#include <chrono>
#include <thread>

// Function to simulate sending data
void send_data(zmq::socket_t& publisher, const std::string& message) {
    zmq::message_t zmq_message(message.data(), message.size());
    publisher.send(zmq_message, zmq::send_flags::none);
    std::cout << "Sent: " << message << std::endl;
}

int main() {
    zmq::context_t context(1);  // 1 I/O thread
    zmq::socket_t publisher(context, zmq::socket_type::pub);

    // Bind the publisher to a TCP socket
    publisher.bind("tcp://*:5555");  // You can change the port as needed

    // Simulate data publishing in a loop
    int message_count = 0;

    while (true) {
        std::string message = "Message " + std::to_string(message_count++);
        send_data(publisher, message);

        // Sleep for a while to simulate data production
        std::this_thread::sleep_for(std::chrono::milliseconds(1000));  // Send a message every second
    }

    return 0;
}
