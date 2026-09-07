# GatewayTecxol-Driver

Driver/gateway en Go que lee inversores (SunSpec v0.1) y publica telemetría al **Fog** vía **MQTT 5 sobre TLS**.

## Sprint 0 — Estado

Pipeline end-to-end reproducible:
- Binario Go estático, sin dependencias C.
- Imagen Docker multi-arch `linux/arm64,linux/amd64`.
- Lectura SunSpec sobre TCP Modbus (modelos 101/103/120/121/123).
- Publicación MQTT 5 con TLS mutuo y LWT.
- Endpoint `/healthz` y `/metrics` Prometheus.
- Apagado limpio bajo `SIGTERM`.

## Quickstart (desarrollo)

```bash
docker compose -f docker-compose.dev.yml up --build
```

Servicios:
- `mosquitto` → `localhost:1883`
- `sunspec-simulator` → `localhost:5020`
- `gateway` → `http://localhost:9091/metrics` y `http://localhost:9091/healthz`

Suscripción rápida a la telemetría:
```bash
docker compose -f docker-compose.dev.yml exec mosquitto \
  mosquitto_sub -h localhost -t 'tecxol/dev01/telemetry' -v
```

## Build local

```bash
make build
./bin/gateway run --config config.example.yaml
```

## Diagnóstico

```bash
./bin/gateway doctor --config config.example.yaml
./bin/gateway version
```

## Estructura

```
cmd/gateway/          # entrypoint CLI (cobra)
internal/config/      # config YAML + env (viper)
internal/observability/ # logger estructurado
internal/inverter/    # interfaz Driver + registry
internal/inverter/sunspec/  # cliente Modbus SunSpec
internal/mqtt/        # cliente MQTT 5 con TLS
internal/health/      # /healthz + /metrics
deploy/systemd/       # unit para instalación nativa
docs/                 # arquitectura y contrato
```

## Documentación

- [docs/architecture.md](docs/architecture.md)
- [docs/fog-contract.md](docs/fog-contract.md)
- [docs/sunspec.md](docs/sunspec.md)

## Roadmap

Ver [docs/roadmap.md](docs/roadmap.md) o el plan de Sprint 0 en la descripción del repo.
