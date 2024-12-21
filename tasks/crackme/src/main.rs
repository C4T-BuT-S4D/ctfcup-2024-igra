use std::cmp;
use std::io;

use goldberg::goldberg_int;
use goldberg::goldberg_stmts;

fn hash(s: &[u8]) -> u64 {
    goldberg_stmts! {
        let mut h = 0u64;

        for c in s.iter() {
            h = (h ^ (*c as u64)).wrapping_mul(31337);
        }

        h
    }
}

// E2DAE8C479AEE65BFFBC0DA49C195C99

fn check_flag(flag: &str) -> bool {
    if flag.len() != 32 {
        return false;
    }

    let target = [
        goldberg_int! {11200688220910254682u64},
        goldberg_int! {11200380368408825194u64},
        goldberg_int! {16146349676620574858u64},
        goldberg_int! {11200565025114669588u64},
        goldberg_int! {12165273152048139049u64},
        goldberg_int! {9396870039570576529u64},
        goldberg_int! {18071888434317316754u64},
        goldberg_int! {14218471961505630534u64},
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
