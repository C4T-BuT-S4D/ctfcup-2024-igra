#include "seccomp-bpf.h"
#include <unistd.h>

static int install_syscall_filter(void) {
  struct sock_filter filter[] = {
      VALIDATE_ARCHITECTURE, EXAMINE_SYSCALL,     ALLOW_SYSCALL(exit_group),
      ALLOW_SYSCALL(exit),   ALLOW_SYSCALL(read), ALLOW_SYSCALL(write),
      KILL_PROCESS,
  };
  struct sock_fprog prog = {
      .len = (unsigned short)(sizeof(filter) / sizeof(filter[0])),
      .filter = filter,
  };

  if (prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)) {
    return 1;
  }
  if (prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, &prog)) {
    return 1;
  }
  return 0;
}

int main() {
  char buf[16];

  if (install_syscall_filter()) {
    return 17;
  }

  read(0, buf, 1024);
}
