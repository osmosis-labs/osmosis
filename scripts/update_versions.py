import re
import networkx as nx
import subprocess

def main():
    # Get the paths to the go modules
    paths = parse_go_work_paths()
    path_to_name = {path: parse_go_mod_name(path) for path in paths}
    names = sorted(path_to_name.values())
    names_to_path = {path_to_name[path]: path for path in path_to_name}
    # build dependency graph for our commit order
    G = build_di_graph(names, paths)

    # now we go commit every go mod file in order, replacing dependencies of prior ones.
    sorted_dependencies = list(nx.topological_sort(G))
    for dependency in sorted_dependencies:
      if check_for_diff(names_to_path[dependency]):
        print(dependency)
    print(sorted_dependencies)

def check_for_diff(path):
  # Run the 'git diff' command and store the output in a variable
  output = subprocess.run(['git', 'diff', 'main', '--', path], stdout=subprocess.PIPE)

  # Check the exit status of the command
  if output.returncode == 0:
    # print("No differences found")
    return False
  else:
    # print("Differences found")
    return True

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