mod cpu;
mod meminfo;
mod process;

use meminfo::{MemError, MemInfo};
use std::time::Duration;

fn main() {
    match MemInfo::read() {
        Ok(m) => {
            println!("Total:     {} MB", m.total_kb / 1024);
            println!("Free:      {} MB", m.free_kb / 1024);
            println!("Available: {} MB", m.available_kb / 1024);
            println!("Used:      {} MB", (m.total_kb - m.available_kb) / 1024);
        }
        Err(MemError::Proc(e)) => eprintln!("could not read memory info: {:?}", e),
        Err(MemError::MissingField(name)) => eprintln!("missing field: {}", name),
    }

    match cpu::usage_percent() {
        Ok(p) => println!("CPU usage: {:.1}%", p),
        Err(e) => eprintln!("could not read CPU stats: {:?}", e),
    }
    match process::list() {
    Ok(mut procs) => {
        procs.sort_by(|a, b| b.rss_kb.cmp(&a.rss_kb));
        println!("\n{:>7} {:<16} {:>5} {:>7} {:>9}", "PID", "NAME", "STATE", "THREADS", "RSS(MB)");
        for p in procs.iter().take(10) {
            println!("{:>7} {:<16} {:>5} {:>7} {:>9}", p.pid, p.name, p.state, p.threads, p.rss_kb / 1024);
        }
    }
    Err(e) => eprintln!("could not list processes: {:?}", e),
}
match process::sample_cpu(Duration::from_secs(1)) {
    Ok(mut v) => {
        v.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap());
        println!("\n{:>7} {:<16} {:>6}", "PID", "NAME", "CPU%");
        for (p, pct) in v.iter().take(10) {
            println!("{:>7} {:<16} {:>6.1}", p.pid, p.name, pct);
        }
    }
    Err(e) => eprintln!("could not sample process CPU: {:?}", e),
}
}