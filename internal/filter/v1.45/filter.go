package v145

import "net/http"

func FilterAll(r *http.Request) (bool, string) {
  return true, "All requests are blocked."
}
