#!/usr/bin/env python3
"""Drives real interactive shells through a pty with the spot hook loaded.

For each shell it runs a non-test command, a passing fake test and a failing
one, then checks the state file: two runs, one green, one red, and a face in
$SPOT_FACE. Usage: python3 tests/shell_smoke.py ./spot [zsh bash fish]
"""
import json, os, pty, select, shutil, subprocess, sys, tempfile, time

spot = os.path.abspath(sys.argv[1])
shells = sys.argv[2:] or ["zsh", "bash", "fish"]


def drive(argv, env, lines, timeout=20):
    pid, fd = pty.fork()
    if pid == 0:
        os.execvpe(argv[0], argv, env)
    out = b""

    def read(wait):
        nonlocal out
        end = time.time() + wait
        while time.time() < end:
            r, _, _ = select.select([fd], [], [], 0.05)
            if r:
                try:
                    chunk = os.read(fd, 65536)
                except OSError:
                    return
                if not chunk:
                    return
                out += chunk

    read(1.5)
    for line in lines:
        os.write(fd, (line + "\n").encode())
        read(2.0)
    try:
        os.write(fd, b"exit\n")
    except OSError:
        pass
    read(1.0)
    os.close(fd)
    os.waitpid(pid, 0)
    return out.decode("utf-8", "replace")


failed = False
for sh in shells:
    path = shutil.which(sh)
    if not path:
        print(f"skip {sh}: not installed")
        continue
    home = tempfile.mkdtemp()
    bindir = os.path.join(home, "bin")
    os.mkdir(bindir)
    with open(os.path.join(bindir, "pytest"), "w") as f:
        f.write('#!/bin/sh\nexit "${FAKE_EXIT:-0}"\n')
    os.chmod(os.path.join(bindir, "pytest"), 0o755)
    env = {
        "HOME": home, "PATH": bindir + os.pathsep + os.environ["PATH"], "TERM": "xterm-256color",
        "SPOT_HOME": home, "SPOT_CONFIG": home, "SPOT_REDUCED_MOTION": "1", "LANG": "en_US.UTF-8",
    }
    face_file = os.path.join(home, "face")
    if sh == "zsh":
        rc = os.path.join(home, ".zshrc")
        open(rc, "w").write(f'eval "$({spot} init zsh --prompt)"\n')
        env["ZDOTDIR"] = home
        argv = [path, "-i"]
        show = f'print -rn -- "$SPOT_FACE" > {face_file}'
    elif sh == "bash":
        rc = os.path.join(home, ".bashrc")
        open(rc, "w").write(f'HISTFILE={home}/hist; set -o history; eval "$({spot} init bash --prompt)"\n')
        argv = [path, "--noprofile", "--rcfile", rc, "-i"]
        show = f'printf %s "$SPOT_FACE" > {face_file}'
    else:
        argv = [path, "--no-config", "-i", "-C", f"{spot} init fish --prompt | source"]
        show = f"printf %s $SPOT_FACE > {face_file}"
    log = drive(argv, env, ["true", "pytest -q", "FAKE_EXIT=1 pytest -q" if sh != "fish" else "env FAKE_EXIT=1 pytest -q", show])
    st = json.load(open(os.path.join(home, "state.json"))) if os.path.exists(os.path.join(home, "state.json")) else {}
    face = open(face_file).read() if os.path.exists(face_file) else ""
    ok = st.get("runs") == 2 and st.get("passes") == 1 and st.get("fails") == 1 and "ᴥ" in face
    print(f"{'ok  ' if ok else 'FAIL'} {sh}: runs={st.get('runs')} passes={st.get('passes')} fails={st.get('fails')} face={face!r}")
    if not ok:
        failed = True
        print(log[-3000:])
sys.exit(1 if failed else 0)
