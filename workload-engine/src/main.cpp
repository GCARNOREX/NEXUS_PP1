#include <iostream>
#include <chrono>
#include <cstdlib>
#include <vector>
#include <algorithm>
#include <string>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <queue>
std::mutex queueMutex;
std::condition_variable cv;
std::queue<int> workQueue;
bool workDone = false;

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
void producer(long numItems) {
    for (long i = 0; i < numItems; ++i) {
        {
            std::lock_guard<std::mutex> lock(queueMutex);  // auto-locks, auto-unlocks
            workQueue.push(i);
        }
        cv.notify_one();  // wake up one waiting consumer
    }
    {
        std::lock_guard<std::mutex> lock(queueMutex);
        workDone = true;
    }
    cv.notify_all();  // wake everyone up so they can exit
}

void consumer() {
    while (true) {
        std::unique_lock<std::mutex> lock(queueMutex);
        cv.wait(lock, [] { return !workQueue.empty() || workDone; });  // sleep until there's work

        while (!workQueue.empty()) {
            int item = workQueue.front();
            workQueue.pop();
            lock.unlock();  // release lock while "processing"
            volatile long x = 0;
            for (long j = 0; j < 1000; ++j) x += j;  // simulate tiny bit of work
            lock.lock();  // re-acquire before checking queue again
        }
        if (workDone && workQueue.empty()) break;
    }
}

void runSyncWorkload(long numItems, int numConsumers) {
    std::vector<std::thread> consumers;
    for (int i = 0; i < numConsumers; ++i) {
        consumers.emplace_back(consumer);
    }
    std::thread prod(producer, numItems);
    prod.join();
    for (auto& c : consumers) c.join();
}
int main(int argc, char* argv[]) {
    std::string type = "primes";  // default
    long limit = 1000000;         // default
    int numThreads =4;
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        if (arg.rfind("--threads=", 0) == 0) {
    numThreads = std::atoi(arg.substr(10).c_str());
}
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
    else if (type == "sync") {
    std::cout << "Running sync workload: " << limit << " items, " << numThreads << " consumers...\n";
    runSyncWorkload(limit, numThreads);
    std::cout << "Sync workload complete.\n";
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