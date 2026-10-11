function groupColour(groupNumber) {
  return "\033[" palette[groupNumber % paletteSize + 1] "m"
}

BEGIN {
  FS = ":.*## "
  reset = "\033[0m"
  bold = "\033[1m"
  paletteSize = split("36 32 33 35 34", palette, " ")
  groupCount = 0
  colour = groupColour(0)
  printf "Usage:\n  make %s<target>%s\n", bold, reset
}

/^##@ / {
  colour = groupColour(groupCount++)
  printf "\n%s%s%s%s\n", bold, colour, substr($0, 5), reset
  next
}

/^[a-zA-Z0-9_-]+:.*## / {
  printf "  %s%-20s%s %s\n", colour, $1, reset, $2
}
