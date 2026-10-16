# Go marks a generated file with this line before its package clause.
function isGenerated(path,    line) {
  while ((getline line < path) > 0) {
    if (line ~ /^\/\/ Code generated .* DO NOT EDIT\.$/) {
      close(path)
      return 1
    }
    if (line ~ /^package /) {
      break
    }
  }
  close(path)
  return 0
}

NR == 1 {
  print
  next
}

{
  path = sourcePath($1)
  if (!(path in generated)) {
    generated[path] = isGenerated(path)
  }
  if (!generated[path]) {
    print
  }
}
