import requests
import re

def fetch_go_mod_from_tag(tag):
    url = f"https://raw.githubusercontent.com/osmosis-labs/osmosis/{tag}/go.mod"
    try:
        response = requests.get(url)
        response.raise_for_status()
        go_mod = response.text
        return go_mod
    except requests.exceptions.RequestException as e:
        raise Exception(f"An error occurred while fetching data from {url}: {e}")


def extract_specific_package_versions(go_mod_content, package_name):
    # Regular expression to find lines containing package names and versions in require section
    require_pattern = r'^\s*(github.com/[^\s]+)\s+([^\s]+)\s*$'

    # Regular expression to find lines containing package names and versions in replace section
    replace_pattern = r'^\s*github.com/([^/]+/[^/]+)\s+=>\s+github.com/([^/]+/[^/]+)\s+([^\s]+)\s*$'

    # Find all matches in the require section using the regular expression
    require_matches = re.findall(require_pattern, go_mod_content, re.MULTILINE)

    # Check if the package_name has a replace in the replace section
    replace_match = re.search(replace_pattern.replace('package_name', package_name), go_mod_content, re.MULTILINE)

    # If there is a match in the replace section, return the replace package_name version
    if replace_match:
        return f"{replace_match.group(2)}@{replace_match.group(3)}"
    
    # If there is no match in the replace section, check the require section for the package_name version
    for match in require_matches:
        if match[0].endswith(package_name):
            return match[1]

    # If the package_name is not found in the require and replace sections, return None
    return None


def get_package_require_version(go_mod, package_name):
    # Find package version in the "require" block only
    require_block = re.search(r'require \((.*?)\)', go_mod, re.DOTALL)
    
    # If "require" block is found
    if require_block:
        # Define regex pattern for version extraction
        version_pattern = re.compile(package_name + r' v([a-zA-Z0-9.\-]+)')
        version = version_pattern.search(require_block.group(1))
        
        # If version is found
        if version:
            return version.group(1)
    
    # Return None if no match found
    return None


def get_package_replace_version(go_mod, package_name):
    # Find package version in the "replace" block only
    replace_block = re.search(r'replace \((.*?)\)', go_mod, re.DOTALL)

    if replace_block:
        # Define regex pattern for version extraction
        replace_pattern = re.compile(package_name + r' => (\S+) v([a-zA-Z0-9.\-]+)')
        replace = replace_pattern.search(replace_block.group(1))

        if replace:
            return replace.group(1).replace("github.com/", "") + "@" + replace.group(2)

    return None

def get_package_version(go_mod, package_name):
    return get_package_replace_version(go_mod, package_name) or get_package_require_version(go_mod, package_name)
