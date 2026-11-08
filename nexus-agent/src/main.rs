mod meminfo;

use meminfo::{MemError, MemInfo};

fn main() {
    match MemInfo::read() {
        Ok(m) => {
            let used_kb = m.total_kb - m.available_kb;
            println!("Total:     {} MB", m.total_kb / 1024);
            println!("Free:      {} MB", m.free_kb / 1024);
            println!("Available: {} MB", m.available_kb / 1024);
            println!("Used:      {} MB", used_kb / 1024);
        }
        Err(MemError::Io(e)) => eprintln!("could not read /proc/meminfo: {}", e),
        Err(MemError::MissingField(name)) => eprintln!("missing field in /proc/meminfo: {}", name),
    }
}