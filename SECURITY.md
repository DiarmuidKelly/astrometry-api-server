# Security Policy

## Current Security Status

**⚠️ This project is currently designed for LOCAL/DEVELOPMENT use only.**

The containerized deployment architecture has a known security limitation that makes it unsuitable for public internet deployment without additional security layers.

## Vulnerability: Docker Socket Access

### The Issue

The containerized deployment requires mounting the Docker socket (`/var/run/docker.sock`) into the API server container to spawn astrometry solver containers. This grants the container **root-equivalent access to the host system**.

### Attack Scenario

If an attacker compromises the API server container (e.g., through a vulnerability in the API code, dependencies, or uploaded image processing), they can:

1. Execute arbitrary Docker commands on the host
2. Spawn privileged containers: `docker run --privileged -v /:/host ...`
3. Mount the host filesystem and gain full root access
4. Access other containers on the same host
5. Pivot to complete host compromise

**Severity:** Critical (CVSS Base Score: 9.0+)

### Why This Design?

Astrometry.net plate-solving requires running the `solve-field` command, which is only available in Docker containers. To make the API server work in a containerized environment, it needs to spawn these solver containers, which requires Docker socket access.

## Safe Use Cases

This deployment is **safe** for:

✅ **Local development** - Running on your own machine for testing
✅ **Trusted internal networks** - Laboratory or observatory environments with network segmentation
✅ **Single-user systems** - Where you control all inputs and trust all users
✅ **Air-gapped environments** - No external network access

## Unsafe Use Cases

This deployment is **NOT safe** for:

❌ **Public internet deployment** - Directly exposed to untrusted users
❌ **Multi-tenant environments** - Shared hosting or cloud platforms
❌ **Untrusted networks** - Any environment where attackers could reach the API
❌ **Production services** - Without additional security layers (see below)

## Mitigation Strategies

### Short Term (For Testing/Development)

1. **Network Isolation**
   - Use firewall rules to restrict API access to trusted IPs only
   - Deploy behind VPN
   - Use SSH tunneling for remote access

2. **Input Validation**
   - Only upload images from trusted sources
   - Implement file type validation and size limits (already in place)
   - Consider scanning uploaded files for malicious content

3. **Monitoring**
   - Monitor Docker API calls from the container
   - Set up alerts for suspicious container spawning
   - Log all API requests

### Long Term (For Production)

#### Option 1: Bare-Metal API Server (Recommended for Small Deployments)

Run the API server directly on the host (not in a container):

```bash
# On host machine
export ASTROMETRY_INDEX_PATH=/path/to/indexes
./astrometry-api-server
```

**Pros:**
- ✅ No docker socket mount needed
- ✅ Uses host's Docker directly (proper isolation)
- ✅ Simple to deploy

**Cons:**
- ❌ Not containerized
- ❌ Requires Docker on host

#### Option 2: Dedicated Solver Service (Recommended for Production)

Separate the web-facing API from the Docker-enabled solver:

```
Internet → Web API (no Docker access)
              ↓ Internal HTTP
         Solver Service (has Docker access)
              ↓
         Spawns solver containers
```

**Architecture:**

1. **Web API Service** (public-facing)
   - Handles HTTP requests
   - Input validation
   - Authentication
   - NO Docker socket access

2. **Internal Solver Service** (private network only)
   - Receives solve requests via internal HTTP/gRPC
   - Has Docker socket access
   - NOT exposed to internet

**Security improvements:**
- Attacker must compromise TWO services to reach host
- Network segmentation adds defense-in-depth
- Web API can be hardened independently
- Solver service isolated from direct internet access

**Implementation:** See [Issue #XXX](../../issues/XXX) for planned architecture

#### Option 3: Docker Socket Proxy

Use a restricted Docker socket proxy like [tecnativa/docker-socket-proxy](https://github.com/Tecnativa/docker-socket-proxy):

```yaml
services:
  docker-proxy:
    image: tecnativa/docker-socket-proxy
    environment:
      CONTAINERS: 1   # Allow container operations
      EXEC: 1         # Allow exec (needed for solver)
      POST: 0         # Deny dangerous operations
      IMAGES: 0       # Deny image operations
      VOLUMES: 0      # Deny volume operations
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  api:
    environment:
      DOCKER_HOST: tcp://docker-proxy:2375
```

**Pros:**
- ✅ Limits what API can do with Docker
- ✅ Still containerized
- ✅ Relatively simple to add

**Cons:**
- ❌ Additional complexity
- ❌ Still some risk if proxy is misconfigured
- ❌ May need custom configuration for solve operations

#### Option 4: Kubernetes with Pod Security Policies

For Kubernetes deployments:

1. Use separate solver pods with `hostPath` Docker socket mount
2. Web API pods have NO special permissions
3. Communication via internal Service
4. Network policies to isolate solver pods
5. Pod Security Policies to prevent privilege escalation

## Reporting a Vulnerability

If you discover a security vulnerability in this project:

1. **Do NOT** open a public GitHub issue
2. Email security details to: [your-security-email@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours and work with you on a fix.

## Roadmap

We are actively working on a production-ready architecture. See:

- [Issue #XXX](../../issues/XXX) - Dedicated Solver Service Architecture
- [Issue #YYY](../../issues/YYY) - Docker Socket Proxy Integration
- [Discussion #ZZZ](../../discussions/ZZZ) - Production Deployment Patterns

## References

- [Docker Socket Security](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [OWASP Container Security](https://owasp.org/www-project-docker-top-10/)

## Acknowledgments

Thank you to everyone who has provided security feedback and suggestions for improving this project's security posture.

---

**Last Updated:** 2025-12-15
**Severity Assessment:** Critical for public deployment, Acceptable for local/trusted use
