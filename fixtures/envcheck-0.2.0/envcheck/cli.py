"""envcheck: verify that the environment variables a project needs are set.

Usage: envcheck [--rc PATH]

Reads .envcheckrc from the current directory (or --rc PATH), then reports
which required variables are missing from the environment.

Exit codes: 0 all present, 1 something missing, 2 the rc file is unusable.
"""

import os
import sys

from envcheck.rc import RcError, parse

RC_DEFAULT = ".envcheckrc"


def main(argv=None):
    argv = list(sys.argv[1:] if argv is None else argv)

    rc_path = RC_DEFAULT
    if argv and argv[0] == "--rc":
        if len(argv) < 2:
            sys.stderr.write("envcheck: --rc needs a path\n")
            return 2
        rc_path = argv[1]
        argv = argv[2:]
    if argv:
        sys.stderr.write("envcheck: unexpected argument %r\n" % argv[0])
        return 2

    try:
        with open(rc_path) as handle:
            text = handle.read()
    except OSError as error:
        sys.stderr.write("envcheck: cannot read %s: %s\n" % (rc_path, error.strerror))
        return 2

    try:
        names = parse(text)
    except RcError as error:
        sys.stderr.write("envcheck: %s\n" % error)
        return 2

    missing = [name for name in names if name not in os.environ]

    for name in missing:
        sys.stdout.write("missing: %s\n" % name)

    if missing:
        sys.stdout.write("%d of %d required variables missing\n" % (len(missing), len(names)))
        return 1

    sys.stdout.write("all %d required variables present\n" % len(names))
    return 0


if __name__ == "__main__":
    sys.exit(main())
