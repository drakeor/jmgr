#include "kernel_info.h"

std::string KernelInfo::getKernelVersion() {
    struct utsname buffer;

    if (uname(&buffer) != 0) {
        return "Error: could not get kernel version.";
    }

    return buffer.release;
}
