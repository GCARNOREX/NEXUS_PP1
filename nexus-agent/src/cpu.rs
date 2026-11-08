use procfs::{CurrentSI, KernelStats, ProcError};
use std::{thread, time::Duration};

struct Sample {
    busy: u64,
    total: u64,
}

fn sample() -> Result<Sample, ProcError> {
    let t = KernelStats::current()?.total;
    let idle = t.idle + t.iowait.unwrap_or(0);
    let total = t.user + t.nice + t.system + idle
        + t.irq.unwrap_or(0) + t.softirq.unwrap_or(0) + t.steal.unwrap_or(0);
    Ok(Sample { busy: total - idle, total })
}

pub fn usage_percent() -> Result<f64, ProcError> {
    let a = sample()?;
    thread::sleep(Duration::from_secs(1));
    let b = sample()?;
    let dt = (b.total - a.total) as f64;
    if dt == 0.0 {
        return Ok(0.0);
    }
    Ok((b.busy - a.busy) as f64 / dt * 100.0)
}