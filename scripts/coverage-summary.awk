function ratio(covered, statements) {
  return statements ? 100 * covered / statements : 0
}

function percent(covered, statements) {
  return sprintf("%.1f%%", ratio(covered, statements))
}

NR == 1 {
  next
}

{
  package = sourcePath($1)
  sub(/\/[^\/]*$/, "", package)
  statements[package] += $2
  totalStatements += $2
  if ($3 > 0) {
    covered[package] += $2
    totalCovered += $2
  }
}

END {
  print "## Go coverage"
  print ""
  printf "**%s** of the statements, generated code left out.\n", percent(totalCovered, totalStatements)
  print ""
  if (ratio(totalCovered, totalStatements) < target) {
    print "> [!WARNING]"
    printf "> Below the soft target of %s%%.\n", target
    print ""
  }
  print "| Package | Coverage |"
  print "| --- | ---: |"
  fflush()
  for (package in statements) {
    printf "| `%s` | %s |\n", package, percent(covered[package], statements[package]) | "sort"
  }
  close("sort")
}
