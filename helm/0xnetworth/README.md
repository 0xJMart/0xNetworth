# 0xNetworth Helm Chart

Helm chart for deploying 0xNetworth Investment Tracking Dashboard to Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Access to GitHub Container Registry (GHCR) for images

## Installation

### From OCI Registry

```bash
helm install 0xnetworth oci://ghcr.io/0xjmart/0xnetworth/0xnetworth --version 0.1.0
```

### From Local Chart

```bash
helm install 0xnetworth ./helm/0xnetworth
```

## Configuration

### Required Configuration

Before deploying, you need to create a Kubernetes secret with your Coinbase API credentials:

```bash
kubectl create secret generic 0xnetworth-secrets \
  --from-literal=COINBASE_API_KEY=your-api-key \
  --from-literal=COINBASE_API_SECRET=your-api-secret \
  -n 0xnetworth
```

Then enable the secret in values.yaml:

```yaml
secret:
  create: false  # Set to false since we created it manually
  name: "0xnetworth-secrets"
```

Or use Sealed Secrets for GitOps workflows.

### Values

Key configuration options in `values.yaml`:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `backend.image.repository` | Backend image repository | `ghcr.io/0xjmart/0xnetworth/backend` |
| `backend.image.tag` | Backend image tag | `latest` |
| `frontend.image.repository` | Frontend image repository | `ghcr.io/0xjmart/0xnetworth/frontend` |
| `frontend.image.tag` | Frontend image tag | `latest` |
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `traefik` |
| `namespace.create` | Create namespace | `true` |
| `namespace.name` | Namespace name | `0xnetworth` |

### Ingress Configuration

To enable ingress, set `ingress.enabled: true` and configure:

```yaml
ingress:
  enabled: true
  className: "traefik"
  hosts:
    - host: 0xnetworth.local
      paths:
        - path: /
          pathType: Prefix
  tls: []
```

## Resource Limits

Default resource limits are configured for Raspberry Pi homelab environments:

- Backend: 100m CPU request, 500m limit; 128Mi memory request, 512Mi limit
- Frontend: 50m CPU request, 200m limit; 64Mi memory request, 256Mi limit

Adjust in `values.yaml` as needed.

## Security

The chart includes security best practices:

- Non-root user execution
- Dropped capabilities
- Seccomp profiles
- Security contexts

API keys should be stored in Kubernetes secrets, not in values.yaml.

## Uninstallation

```bash
helm uninstall 0xnetworth
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n 0xnetworth
```

### View Logs

```bash
# Backend logs
kubectl logs -n 0xnetworth deployment/0xnetworth-backend

# Frontend logs
kubectl logs -n 0xnetworth deployment/0xnetworth-frontend
```

### Verify Secret

```bash
kubectl get secret 0xnetworth-secrets -n 0xnetworth -o yaml
```

### Test API

```bash
kubectl port-forward -n 0xnetworth svc/0xnetworth-backend 8080:8080
curl http://localhost:8080/api/health
```

