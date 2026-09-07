# Arquitectura (Sprint 0)

```
+------------------+      TCP/502       +------------------+
|  Inversor        | <----------------> |  Gateway Driver  |
|  (SunSpec Modbus)|                    |  (Go, este repo) |
+------------------+                    +---------+--------+
                                                  |
                                          TLS 8883| (MQTT 5)
                                                  v
                                        +---------+--------+
                                        |       Fog        |
                                        |  (broker MQTT 5) |
                                        +------------------+
```

## Límites del Sprint 0

- **Dentro:** 1 inversor SunSpec por gateway, polling, telemetría, métricas, health, Docker multi-arch, CI.
- **Fuera:** drivers propietarios, multi-inversor, búfer offline, OTA, autenticación avanzada.

## Componentes

- `cmd/gateway` — entrypoint (subcomandos `run`, `version`, `doctor`).
- `internal/inverter.Driver` — interfaz común (`Scan(ctx) -> Snapshot`).
- `internal/inverter/sunspec` — implementación sobre Modbus TCP, lectura de base y modelos comunes.
- `internal/mqtt` — wrapper sobre `eclipse/paho.golang` con TLS y reconexión.
- `internal/health` — estado del servicio, `/healthz` y `/metrics` Prometheus.
- `internal/observability` — logger `zap` (json/console).

## Configuración

YAML + override por env (`GATEWAY_<KEY>__<SUB>=value`). Ver `config.example.yaml`.
