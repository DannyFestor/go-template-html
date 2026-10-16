BEGIN {
  while ((getline line < "go.mod") > 0) {
    if (split(line, words, " ") == 2 && words[1] == "module") {
      module = words[2]
      break
    }
  }
  close("go.mod")
}

function sourcePath(block) {
  return substr(block, length(module) + 2, index(block, ":") - length(module) - 2)
}
