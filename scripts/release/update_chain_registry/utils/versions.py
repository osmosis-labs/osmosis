import re

def parse_version(version):

    major, minor, patch = map(int, version[1:].split('.'))
    return major, minor, patch

def same_major(version_1, version_2):
    # Parse the major versions
    major_1, _, _ = parse_version(version_1)
    major_2, _, _ = parse_version(version_2)

    # Compare the major versions
    return major_1 == major_2

def compare_versions(version_1, version_2):
    # Parse the versions into major, minor, patch
    major_1, minor_1, patch1 = parse_version(version_1)
    major_2, minor_2, patch2 = parse_version(version_2)

    # Compare the major versions
    if major_1 > major_2:
        return 1
    elif major_1 < major_2:
        return -1

    # If major versions are equal, compare the minor versions
    if minor_1 > minor_2:
        return 1
    elif minor_1 < minor_2:
        return -1

    # If minor versions are equal, compare the patch versions
    if patch1 > patch2:
        return 1
    elif patch1 < patch2:
        return -1

    # If all parts are equal, the versions are the same
    return 0

def validate_tag(tag):
    pattern = '^v[0-9]+.[0-9]+.[0-9]+$'
    return bool(re.match(pattern, tag))
