#include <iostream>
#include <sstream>
#include <cstdio>
#include <string>
#include <array>

class GPUInfo {
public:
    // Method to get GPU usage by calling `nvidia-smi`
    void getNvidiaGPUUsage() {
        const char* cmd = "nvidia-smi --query-gpu=utilization.gpu,utilization.memory,memory.total,memory.used --format=csv,noheader,nounits";
        std::array<char, 128> buffer;
        std::string result = "";

        // Open a pipe to run the command
        FILE* pipe = popen(cmd, "r");
        if (!pipe) {
            std::cerr << "Error: could not run `nvidia-smi`." << std::endl;
            return;
        }

        // Read the output of `nvidia-smi`
        while (fgets(buffer.data(), buffer.size(), pipe) != nullptr) {
            result += buffer.data();
        }

        // Close the pipe
        pclose(pipe);

        // Parse and print the result
        std::cout << "Nvidia GPU Usage:" << std::endl;
        std::cout << "------------------------" << std::endl;
        parseAndPrintGPUUsage(result);
    }

private:
    // Method to parse and print GPU usage information
    void parseAndPrintGPUUsage(const std::string& result) {
        std::istringstream iss(result);
        std::string gpuUsage, memoryUsage, totalMemory, usedMemory;

        // Read CSV output
        while (iss >> gpuUsage >> memoryUsage >> totalMemory >> usedMemory) {
            std::cout << "GPU Utilization: " << gpuUsage << "%" << std::endl;
            std::cout << "Memory Utilization: " << memoryUsage << "%" << std::endl;
            std::cout << "Total Memory: " << totalMemory << " MiB" << std::endl;
            std::cout << "Used Memory: " << usedMemory << " MiB" << std::endl;
            std::cout << "------------------------" << std::endl;
        }
    }
};

int main() {
    GPUInfo gpuInfo;

    // Get Nvidia GPU usage
    gpuInfo.getNvidiaGPUUsage();

    return 0;
}
