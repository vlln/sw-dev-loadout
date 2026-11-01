"""Parse .envcheckrc: one required environment variable name per line."""


class RcError(Exception):
    pass


def parse(text):
    """Return the list of required variable names, in file order.

    Blank lines are skipped. A line starting with '#' is a comment.
    Duplicate names keep their first position.
    """
    names = []
    for lineno, raw in enumerate(text.splitlines(), start=1):
        line = raw.strip()
        if not line:
            continue
        if line.startswith("#"):
            continue
        if " " in line:
            raise RcError("line %d: %r contains a space" % (lineno, line))
        if line not in names:
            names.append(line)
    return names
