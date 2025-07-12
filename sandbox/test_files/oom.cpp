#include <vector>
#include <cstddef>
#include <iostream>

int main() {
    std::vector<char*> allocations;
    const std::size_t chunkSize = 1024 * 1024; // 1 MiB per allocation

    while (true) {
        char* p = new (std::nothrow) char[chunkSize];
        if (!p) {
            break;
        }
        allocations.push_back(p);
    }

    return 0;
}
