package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"zeus/internal/filter"

	"github.com/sirupsen/logrus"
)

const (
  DefaultDockerSocketPath = "/var/run/docker.sock"
  EnvDockerSocketPath = "DOCKER_SOCKET_PATH"
)



type server struct {
  proxy *httputil.ReverseProxy
  listener net.Listener
}

func (self *server) interceptor(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/_ping" {
      logrus.Infoln("Ignore filters for ping request")
      self.proxy.ServeHTTP(w, r)
      return
    }

    if block, msg := filter.ShouldBlock(r); block {
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusMethodNotAllowed)

      json.NewEncoder(w).Encode(map[string]string{"message": msg})
      logrus.Infof("Blocked Request %s %s\n", r.Method, r.URL.Path)
      return 
    }

    logrus.Infof("Forward Request %s %s\n", r.Method, r.URL.Path)
    self.proxy.ServeHTTP(w, r)
}

func (self *server) run() {
  defer self.listener.Close()
  if err := http.Serve(
    self.listener,
    http.HandlerFunc(self.interceptor),
  ); err != nil {
    panic(err)
  }
}

func NewServer(
  dockerSocketPath string,
) server {
  proxy := &httputil.ReverseProxy{
    Director: func(req *http.Request) {
      req.URL.Scheme = "http"
      req.URL.Host = "docker"
    },
    Transport: &http.Transport{
      Dial: func(network, addr string) (net.Conn, error) {
        return net.Dial("unix", dockerSocketPath)
      },
    },
  }

  listener, err := net.FileListener(os.NewFile(3, "systemd socket"))
  if err != nil {
    panic(err)
  }

  return server{
    proxy: proxy,
    listener: listener,
  }
}

func main() {
  logrus.SetFormatter(
    &logrus.TextFormatter{
      TimestampFormat: "2006-01-02 15:04:05",
      FullTimestamp: true,
    },
  )

  dockerSocketPath := DefaultDockerSocketPath
  if path := os.Getenv(EnvDockerSocketPath); path != "" {
    dockerSocketPath = path
  }

  s := NewServer(dockerSocketPath)
  s.run()
}


