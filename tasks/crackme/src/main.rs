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
        5164074131756060218u64,
        5163766467803092330u64,
        16146349645197335434u64,
        5163951155930170580u64,
        6127674572811815721u64,
        9395885140971293393u64,
        18070903755689734642u64,
        14217487282878048422u64,
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
