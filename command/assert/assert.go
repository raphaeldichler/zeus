package assert

func True(b bool, msg string) {
  if !b {
    panic(msg)
  }
}

func Unreachable(msg string) {
  panic(msg)
}
