# Ingress

Ingress is the component responsible for managing and routing external traffic into internal services, typically using defined rules and configurations. It acts as the entry point to your application or service within a containerized environment.

## Errors

The ingress daemon can encounter errors due to various reasons, many of which are outside the control of the daemon itself. We categorize these errors into three groups: `ingress`, `server`, and `tls`.

### ingress

Errors in this category occur when the ingress daemon is unable to create the necessary resources required to function properly. Specifically, this includes failures in creating or running the ingress container. In such cases, an error labeled `FailedCreatingIngressContainer` will appear.

When this happens, the system will apply an automatic recovery strategy that retries the operation until the container is successfully created. In some scenarios, a restart or reload of the Docker daemon may be triggered to resolve the issue.

