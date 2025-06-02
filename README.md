# Zeus
[![Go Report Card](https://goreportcard.com/badge/github.com/raphaeldichler/zeus)](https://goreportcard.com/report/github.com/raphaeldichler/zeus)

<img src="https://github.com/raphaeldichler/zeus/blob/30294adb1b0fad8f4e8632b39dfa939c3bce266e/docs/icon.png" width="100">



----

Zeus is an open-source system for managing containerized applications on a single host. It provides essential mechanisms for deploying and maintaining applications.

Zeus offers an alternative to [Kubernetes (K8s)](https://kubernetes.io/) for smaller applications that do not require multi-host deployment for performance or stability. While some features of K8s are useful at any scale, others are unnecessary for simpler setups. Zeus aims to deliver the core benefits of K8s in an environment suitable for the majority of applications that do not need a full-blown cluster to serve users.

Zeus provides a declarative way to define the desired state of your application, inspired by Kubernetes. Additionally, it encourages a development approach where applications are designed to run consistently across environments.

----

## Why zeus?

Zeus was born from the observation that many applications face similar requirements and challenges. We’ve also seen developers adopting K8s in scenarios where it might not be the most appropriate solution. While we acknowledge that Kubernetes offers many powerful features that make it appealing, we believe that most use cases only require a subset of these capabilities. We also recognize that designing systems with scalability in mind is a common aspiration, and Zeus enables this future growth without overcomplicating the present.

### Key Features
 - Ingress controller with TLS encryption
 - Service-like interaction between containers within the same Zeus application
 - Zero-downtime updates
 - Secret management
 - Seamless transition between development and production environments

### The Problem We’re Solving

Let’s consider a scenario involving a monolithic application - often the simplest and least error-prone architectural approach. Initially, deploying a single container may seem sufficient. However, as the application evolves, we may want to observe its behavior and detect problems as early as possible.

One option could be integrating custom observability features directly into the monolith, but this quickly becomes overkill. A more practical solution is to run a [Grafana](https://grafana.com/) container alongside the application to visualize metrics. To feed data into Grafana, we also need a [Prometheus](https://prometheus.io/) container to collect metrics from the application.

What started as a simple, single-container deployment now requires three containers just to enable basic observability.

Beyond monitoring, we also need a way to interact with the application externally. This typically leads to introducing a reverse proxy to coordinate access between services. And if we want to serve the application securely over HTTPS with a trusted certificate, we’ll need a container to handle certificate generation and renewal.

In the end, we’ve gone from a single container to a small but complex setup involving multiple services, each of which must be configured, maintained, and deployed reliably in a production environment.
