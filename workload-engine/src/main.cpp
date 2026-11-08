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
void runSortWorkload(long size) {
    std::vector<int> data(size);
    for (long i = 0; i < size; ++i) {
        data[i] = std::rand();
    }
    std::sort(data.begin(), data.end());
}

// Multiplies two N×N matrices of doubles using plain nested loops.
void runMatrixWorkload(long n) {
    std::vector<std::vector<double>> A(n, std::vector<double>(n));
    std::vector<std::vector<double>> B(n, std::vector<double>(n));
    std::vector<std::vector<double>> C(n, std::vector<double>(n, 0.0));

    // Fill A and B with random values
    for (long i = 0; i < n; ++i) {
        for (long j = 0; j < n; ++j) {
            A[i][j] = static_cast<double>(std::rand()) / RAND_MAX;
            B[i][j] = static_cast<double>(std::rand()) / RAND_MAX;
        }
    }

    // C = A * B
    for (long i = 0; i < n; ++i) {
        for (long j = 0; j < n; ++j) {
            double sum = 0.0;
            for (long k = 0; k < n; ++k) {
                sum += A[i][k] * B[k][j];
            }
            C[i][j] = sum;
        }
    }
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
    } 
    else if (type == "matrix") {
    std::cout << "Multiplying " << limit << "x" << limit << " matrices...\n";
    runMatrixWorkload(limit);
    std::cout << "Matrix multiplication complete.\n";
    }
    else {
        std::cout << "Unknown type: " << type << "\n";
        return 1;
    }

    auto end = std::chrono::high_resolution_clock::now();
    std::chrono::duration<double> elapsed = end - start;
    std::cout << "Time taken: " << elapsed.count() << " seconds\n";

    return 0;
}