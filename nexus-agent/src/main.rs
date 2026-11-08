mod cpu;
mod meminfo;

use meminfo::{MemError, MemInfo};

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
}