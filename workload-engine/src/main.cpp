#include <iostream>
#include <chrono>
#include <cstdlib>
#include <vector>
#include <algorithm>
#include <string>

// Counts primes below `limit` using simple trial division.
long countPrimesBelow(long limit) {
    long count = 0;
    for (long n = 2; n < limit; ++n) {
        bool isPrime = true;
        for (long d = 2; d * d <= n; ++d) {
            if (n % d == 0) {
                isPrime = false;
                break;
            }
        }
        if (isPrime) count++;
    }
    return count;
}

int main(int argc, char* argv[]) {
    std::string type = "primes";  // default
    long limit = 1000000;         // default

    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        if (arg.rfind("--type=", 0) == 0) {
            type = arg.substr(7);
        } else {
            limit = std::atol(argv[i]);  // plain number = limit/size
        }
    }

    auto start = std::chrono::high_resolution_clock::now();

    if (type == "primes") {
        std::cout << "Counting primes below " << limit << "...\n";
        long result = countPrimesBelow(limit);
        std::cout << "Found " << result << " primes.\n";
    } else if (type == "sort") {
        std::cout << "Sorting " << limit << " random integers...\n";
        runSortWorkload(limit);
        std::cout << "Sort complete.\n";
    } else {
        std::cout << "Unknown type: " << type << "\n";
        return 1;
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double> elapsed = end - start;
    std::cout << "Time taken: " << elapsed.count() << " seconds\n";

    return 0;
}