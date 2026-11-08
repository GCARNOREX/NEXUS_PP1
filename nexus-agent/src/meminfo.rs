use std::{fs, io};

#[derive(Debug)]
pub struct MemInfo {
    pub total_kb: u64,
    pub free_kb: u64,
    pub available_kb: u64,
}

#[derive(Debug)]
pub enum MemError {
    Io(io::Error),
    MissingField(&'static str),
}

impl From<io::Error> for MemError {
    fn from(e: io::Error) -> Self {
        MemError::Io(e)
    }
}

impl MemInfo {
    pub fn read() -> Result<MemInfo, MemError> {
        let text = fs::read_to_string("/proc/meminfo")?;
        Ok(MemInfo {
            total_kb: field(&text, "MemTotal:").ok_or(MemError::MissingField("MemTotal"))?,
            free_kb: field(&text, "MemFree:").ok_or(MemError::MissingField("MemFree"))?,
            available_kb: field(&text, "MemAvailable:").ok_or(MemError::MissingField("MemAvailable"))?,
        })
    }
}

fn field(text: &str, key: &str) -> Option<u64> {
    text.lines()
        .find_map(|l| l.strip_prefix(key)?.split_whitespace().next()?.parse().ok())
}