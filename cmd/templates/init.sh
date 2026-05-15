aworkspace() {
  if [ "$1" = "cd" ]; then
    shift
    cd "$(AWORKSPACE_SHELL=1 command aworkspace cd "$@")"
  else
    command aworkspace "$@"
  fi
}