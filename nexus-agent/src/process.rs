use procfs::process::all_processes;
use procfs::ProcError;
use std::{collections::HashMap, thread, time::Duration};
use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct ProcInfo {
    pub pid: i32,
    pub name: String,
    pub state: char,
    pub threads: i64,
    pub rss_kb: u64,
    pub cpu_ticks: u64, // utime + stime
}

pub fn list() -> Result<Vec<ProcInfo>, ProcError> {
    let page_kb = procfs::page_size() / 1024;
    let mut out = Vec::new();
    for entry in all_processes()? {
        // A process can exit while we're reading it, so skip failures instead of using `?`
        let p = match entry {
            Ok(p) => p,
            Err(_) => continue,
        };
        let s = match p.stat() {
            Ok(s) => s,
            Err(_) => continue,
        };
        out.push(ProcInfo {
            pid: s.pid,
            name: s.comm,
            state: s.state,
            threads: s.num_threads,
            rss_kb: s.rss * page_kb, // rss is counted in pages
            cpu_ticks: s.utime + s.stime,
        });
    }
    Ok(out)
}
pub fn sample_cpu(interval: Duration) -> Result<Vec<(ProcInfo, f64)>, ProcError> {
    let first: HashMap<i32, u64> = list()?.into_iter().map(|p| (p.pid, p.cpu_ticks)).collect();
    thread::sleep(interval);
    let second = list()?;
    let ticks = procfs::ticks_per_second() as f64 * interval.as_secs_f64();
    Ok(second
        .into_iter()
        .filter_map(|p| {
            let before = *first.get(&p.pid)?; // skips processes that started mid-interval
            let pct = p.cpu_ticks.saturating_sub(before) as f64 / ticks * 100.0;
            Some((p, pct))
        })
        .collect())
}
#[derive(Debug, Serialize)]
pub struct ProcSnap {
    #[serde(flatten)]
    pub info: ProcInfo,
    pub cpu_percent: f64,
}

/// Tracks the given PIDs; with no PIDs, returns the 10 busiest processes.
pub fn sample_filtered(interval: Duration, pids: &[i32]) -> Result<Vec<ProcSnap>, ProcError> {
    let mut snaps: Vec<ProcSnap> = sample_cpu(interval)?
        .into_iter()
        .filter(|(p, _)| pids.is_empty() || pids.contains(&p.pid))
        .map(|(info, cpu_percent)| ProcSnap { info, cpu_percent })
        .collect();
    if pids.is_empty() {
        snaps.sort_by(|a, b| b.cpu_percent.partial_cmp(&a.cpu_percent).unwrap());
        snaps.truncate(10);
    }
    Ok(snaps)
}