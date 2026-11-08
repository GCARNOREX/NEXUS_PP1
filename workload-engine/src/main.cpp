#include <iostream>
#include <chrono>
#include <cstdlib>
using namespace std;

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
    long limit = 1000000; // default
    if (argc > 1) {
        limit = atol(argv[1]);
    }

    cout << "Counting primes below " << limit << "...\n";

    auto start = chrono::high_resolution_clock::now();
    long result = countPrimesBelow(limit);
    auto end = chrono::high_resolution_clock::now();

    chrono::duration<double> elapsed = end - start;

    cout << "Found " << result << " primes.\n";
    cout << "Time taken: " << elapsed.count() << " seconds\n";

    return 0;
}