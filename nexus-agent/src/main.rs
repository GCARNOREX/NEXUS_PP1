mod cpu;
mod meminfo;
mod process;

use meminfo::{MemError, MemInfo};
use process::ProcSnap;
use serde::Serialize;
use std::{
    env, fs,
    sync::mpsc,
    thread,
    time::{Duration, SystemTime, UNIX_EPOCH},
};

enum Msg {
    Cpu(f64),
    Mem(MemInfo),
    Procs(Vec<ProcSnap>),
}

#[derive(Serialize)]
struct Snapshot {
    host: String,
    timestamp: u64,
    cpu_percent: f64,
    mem: MemInfo,
    processes: Vec<ProcSnap>,
}

fn hostname() -> String {
    fs::read_to_string("/proc/sys/kernel/hostname")
        .map(|s| s.trim().to_string())
        .unwrap_or_else(|_| "unknown".into())
}

fn now_secs() -> u64 {
    SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()
}

fn main() {
    // PIDs to track come from the command line: cargo run -- 1234 5678
    let pids: Vec<i32> = env::args().skip(1).filter_map(|a| a.parse().ok()).collect();
    let (tx, rx) = mpsc::channel::<Msg>();

    let tx_cpu = tx.clone();
    thread::spawn(move || loop {
        if let Ok(p) = cpu::usage_percent() {
            if tx_cpu.send(Msg::Cpu(p)).is_err() { break; }
        }
    });

   let tx_mem = tx.clone();
thread::spawn(move || loop {
    match MemInfo::read() {
        Ok(m) => {
            if tx_mem.send(Msg::Mem(m)).is_err() { break; }
        }
        Err(MemError::Proc(e)) => eprintln!("mem error: {:?}", e),
        Err(MemError::MissingField(name)) => eprintln!("mem error: missing field {}", name),
    }
    thread::sleep(Duration::from_secs(1));
});
    let tx_proc = tx.clone();
    thread::spawn(move || loop {
        if let Ok(v) = process::sample_filtered(Duration::from_secs(1), &pids) {
            if tx_proc.send(Msg::Procs(v)).is_err() { break; }
        }
    });

    let host = hostname();
    let mut cpu_p: Option<f64> = None;
    let mut mem: Option<MemInfo> = None;

    // The process thread ticks about once a second, so each Procs message triggers one snapshot.
    for msg in rx {
        match msg {
            Msg::Cpu(p) => cpu_p = Some(p),
            Msg::Mem(m) => mem = Some(m),
            Msg::Procs(processes) => {
                if let (Some(c), Some(m)) = (cpu_p, mem.clone()) {
                    let snap = Snapshot {
                        host: host.clone(),
                        timestamp: now_secs(),
                        cpu_percent: c,
                        mem: m,
                        processes,
                    };
                    println!("{}", serde_json::to_string(&snap).unwrap());
                }
            }
        }
    }
}