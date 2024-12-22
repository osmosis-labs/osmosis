import subprocess
import re


def get_commits(branch):
    try:
        # Get the list of commits in the specified branch
        commit_hashes = subprocess.check_output(
            ['git', 'log', '--oneline', 'origin/' + branch]).decode().splitlines()
        return commit_hashes
    except subprocess.CalledProcessError as e:
        print(f"Error: {e}")
        return []


def get_pr_number(commit_message):
    # Extract PR number from commit message
    pr_number = re.search(r'#\d+', commit_message)
    return pr_number.group(0) if pr_number else None


def main():
    release_branch = input("Enter the release branch name: ")
    main_branch = input("Enter the main branch name (default: main): ")

    if not main_branch:
        main_branch = "main"

    # Fetch the latest branches
    subprocess.call(['git', 'fetch'])

    main_commits = get_commits(main_branch)
    release_commits = get_commits(release_branch)

    if not main_commits:
        print(f"No commits found for the {main_branch} branch.")
        return
    if not release_commits:
        print(f"No commits found for the {release_branch} branch.")
        return

    main_pr_numbers = {get_pr_number(commit) for commit in main_commits}
    release_pr_numbers = {get_pr_number(commit) for commit in release_commits}

    # Find commits in the main branch that are not backported to the release branch
    missing_commits = [commit for commit in main_commits if get_pr_number(
        commit) not in release_pr_numbers]

    if not missing_commits:
        print(f"All commits from {main_branch} are present in {release_branch}.")
    else:
        print(f"\nCommits present in {main_branch} but missing in {release_branch}:")
        for commit in missing_commits:
            print(commit)


if __name__ == '__main__':
    main()
