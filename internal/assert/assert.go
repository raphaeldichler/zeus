package assert

func EndsNotWith(str string, suffix rune, msg string) {
  if str[len(str) - 1] == byte(suffix) {
    panic(msg)
  }
}

func Unreachable(msg string) {
  panic(msg)
}

func True(b bool, msg string) {
  if !b {
    panic(msg)
  }
}
