package filter

import (
	"net/http"
	"strings"
	v145 "zeus/internal/filter/v1.45"
)

type (
  ApiVersion string
  RestrictionFilter func(req *http.Request) (bool, string)
)

const (
  ErrMessageVersionNotSpecified = "Request version must be part of the reqeuest"
  ErrMessageUnsupportedVersion = "Request version is not supported"
)

var (
  filters = map[ApiVersion][]RestrictionFilter {
    "v1.45": {
      v145.FilterAll,
    },
  }
)

func ShouldBlock(req *http.Request) (bool, string) {
  hasVersion, version := hasVersion(req)
  if !hasVersion {
    return true, ErrMessageVersionNotSpecified
  }

  filters, ok := filters[version]
  if !ok {
    return true, ErrMessageUnsupportedVersion
  }

  for _, f := range filters {
    if block, msg := f(req); block {
      return true, msg
    }
  }

  return false, ""
}

func hasVersion(r *http.Request) (bool, ApiVersion) {
    parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
    if len(parts) > 0 && strings.HasPrefix(parts[0], "v") {
        return true, ApiVersion(parts[0])
    }
    return false, ""
}
