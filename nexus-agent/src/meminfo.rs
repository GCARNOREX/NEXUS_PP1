use procfs::{Current, Meminfo, ProcError};

#[derive(Debug)]
pub struct MemInfo {
    pub total_kb: u64,
    pub free_kb: u64,
    pub available_kb: u64,
}

#[derive(Debug)]
pub enum MemError {
    Proc(ProcError),
    MissingField(&'static str),
}

impl From<ProcError> for MemError {
    fn from(e: ProcError) -> Self {
        MemError::Proc(e)
    }
}

impl MemInfo {
    pub fn read() -> Result<MemInfo, MemError> {
        let m = Meminfo::current()?;
        let available = m.mem_available.ok_or(MemError::MissingField("MemAvailable"))?;
        Ok(MemInfo {
            total_kb: m.mem_total / 1024,
            free_kb: m.mem_free / 1024,
            available_kb: available / 1024,
        })
    }
}