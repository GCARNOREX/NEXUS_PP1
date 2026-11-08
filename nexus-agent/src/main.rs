use std::fs;
fn main()
{
    let meminfo=fs::read_to_string("/proc/meminfo").expect("cannot read /proc/meminfo");
    for line in meminfo.lines()
    {
        if let Some(kb)=parse_field(line, "MemTotal :")
        {
            println!("MemTotal :{} kB",kb);
        }
    }
}
fn parse_field(line: &str,key: &str)->Option<u64>
{
    let rest = line.strip_prefix(key)?;
    rest.trim().split_whitespace().next()?.parse().ok()

}