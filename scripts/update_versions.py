import re
import networkx as nx
import subprocess
import time

def main():
    # Get the paths to the go modules
    paths = parse_go_work_paths()
      # Create a mapping from the paths to the names of the modules
    path_to_name = {path: parse_go_mod_name(path) for path in paths}
    names = sorted(path_to_name.values())
    names_to_path = {path_to_name[path]: path for path in path_to_name}
    # build dependency graph for our commit order
    G = build_di_graph(names, paths)
    
    # Perform a topological sort on the dependency graph and reverse the resulting list
    sorted_dependencies = list(nx.topological_sort(G))[::-1]

    # now we go commit every go mod file in order, replacing dependencies of prior ones.
    update_dep_versions = {}
    num_commits = 0
    for dependency in sorted_dependencies:
      # see if we need to change anything per update_dep_versions
      mod_path = names_to_path[dependency]
      update_dependencies(mod_path, update_dep_versions)
      if check_for_diff_against_main(mod_path):
        print("detected diff relative to main in " + dependency)
        if check_for_diff_since_last_commit():
          print("detected diff relative to last commit, committing" + dependency)
          subprocess.run("git add .".split(" "), capture_output=True)
          subprocess.run("git commit -m".split(" ") + ["(auto) locking dependency versions for " + dependency])
          num_commits += 1
        v = get_go_sum_version()
        update_dep_versions[dependency] = str(v)

    gitResetStr = "git reset HEAD" + "~"*num_commits
    input("going to run git push, hit enter to confirm. Ctrl-C to cancel." + 
      " (And if so, you likely want to do " + gitResetStr + " to undo the commits)")

    result = subprocess.run("git push".split(" "))
    if result.returncode != 0:
      print("branch isn't already pushed yet, aborting. Push the branch to github, then re-run script.")
      result = subprocess.run(gitResetStr.split(" "))
      return
    
    print("waiting a bit for golang server caches to propogate, sleeping 5 seconds")
    time.sleep(5)

    for dependency in update_dep_versions:
      subprocess.run("go mod tidy".split(" "), cwd=names_to_path[dependency])

def update_dependencies(mod_path, dependencies):
    # Open the go.mod file in read mode
    with open(mod_path + "/go.mod", 'r') as mod_file:
        # Read the contents of the file into a list of lines
        lines = mod_file.readlines()
    changed = False
    # Loop through the lines of the go.mod file
    for i, line in enumerate(lines):
        # Check if the line starts with "require"
        # Split the line by spaces
        parts = line.split(' ')
        if len(parts) < 2: # not a require line
            continue
        # Get the dependency name and version from the line
        name = parts[0].strip()
        version = parts[1].strip()
        rem_parts = ' '.join(parts[2:]).strip()
        # Check if the dependency name is in the dictionary of desired versions
        if name in dependencies:
            human_version = version
            if "-" in human_version:
                human_version = version.split("-")[0]
            new_version = human_version + "-" + dependencies[name]
            # Update the version in the go.mod file
            print(lines[i])
            print(dependencies[name])
            lines[i] = f"\t{name} {new_version} {rem_parts}\n"
            print(lines[i])
            changed = True

    if not changed:
        return
    # Open the go.mod file in write mode
    with open(mod_path + "/go.mod", 'w') as mod_file:
        # Write the modified lines to the file
        mod_file.writelines(lines)

def get_go_sum_version():
  # Define the command as a list of strings
  command = "git --no-pager show --quiet --abbrev=12 --date=format-local:%Y%m%d%H%M%S --format=%cd-%h".split(" ")
  # print(command)
  env = {'TZ': 'UTC'}

  # Run the command
  result = subprocess.run(command, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
  # Print the output
  return result.stdout.decode('utf-8').strip()

def check_for_diff_against_main(path):
  # Run the 'git diff' command and store the output in a variable
  output = subprocess.run(['git', 'diff', 'main', "--exit-code", '--', path], stdout=subprocess.PIPE)
  # Check the exit status of the command (!- 0 means there was a diff)
  return output.returncode != 0

def check_for_diff_since_last_commit():
  output = subprocess.run(['git', 'diff', "--exit-code"], stdout=subprocess.PIPE)
  return output.returncode != 0

def build_di_graph(names, mod_filepaths):
  # Create an empty directed graph
  G = nx.DiGraph() 

  # Iterate over the list of go.mod filepaths
  for filepath in mod_filepaths:
    curname = parse_go_mod_name(filepath)
    # Open the go.mod file in read mode
    with open(filepath + "/go.mod", 'r') as mod_file:
      # Read the file contents into a string
      contents = mod_file.read()
      
      # Use a regular expression to match the 'require' block and extract the dependencies
      dependencies = re.findall(r'require\s+\(\s*(.*?)\s*\)', contents, re.DOTALL)      
  
      # Split the dependencies into a list of individual dependency strings
      dependency_list = dependencies[0].strip().split('\n')
      
      # Add the dependencies as edges in the graph, with the current go.mod file as the source and the dependency as the target
      # print(names)
      for dependency in dependency_list:
        dep = dependency.split(" ")[0].strip()
        # print(dep)
        if dep in names:
          G.add_edge(curname, dep)
  return G

def parse_go_work_paths():
  # Open the .work file in read mode
  with open('go.work', 'r') as work_file:
    # Read the file contents into a string
    contents = work_file.read()
    
    # Use a regular expression to match the 'use' lines and extract the path
    matches = re.findall(r'use (.*)', contents)
    
    # Return the extracted paths
    return matches

def parse_go_mod_name(path):
  # Open the go.mod file in read mode
  with open(path + '/go.mod', 'r') as mod_file:
    # Read the file contents into a string
    contents = mod_file.read()
    
    # Use a regular expression to match the 'module' line and extract the module string
    match = re.search(r'module (.*)', contents)
    
    # Get the module string from the match object
    module_string = match.group(1)
    
    return module_string

main()