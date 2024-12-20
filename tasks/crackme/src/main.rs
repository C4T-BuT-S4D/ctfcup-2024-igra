use std::cmp;
use std::io;

fn hash(s: &[u8]) -> u64 {
    let mut h = 0u64;

    for c in s.iter() {
        h = (h ^ (*c as u64)).wrapping_mul(31337);
    }

    h
}

// e2dae8c479aee65bffbc0da49c195c99

fn check_flag(flag: &str) -> bool {
    if flag.len() != 32 {
        return false;
    }

    let target = [
        11200688220910254682,
        11200380368408825194,
        16146349676620574858,
        11200565025114669588,
        12165273152048139049,
        9396870039570576529,
        18071888434317316754,
        14218471961505630534,
    ];

    let mut i = 0;

    while i < flag.len() {
        let h = hash(flag[i..cmp::min(i + 4, flag.len())].as_bytes());
        if target[i / 4] != h {
            return false;
        }
        i += 4
    }

    true
}

fn main() {
    let mut flag = String::new();
    let stdin = io::stdin();
    stdin.read_line(&mut flag).unwrap();

    if check_flag(flag.trim()) {
        println!("gj")
    } else {
        println!("bj")
    }
}
