# Contrato Fog — MQTT 5 (Sprint 0)

## Conexión
- **Transporte:** MQTT 5 sobre TLS 1.2+.
- **QoS:** 1 para telemetría, 1 retained para health/inventory.
- **Limpia sesión:** `false` (reutiliza subscripciones en reconexión).
- **Reconexión:** exponencial con jitter, tope 5 min.

## Tópicos
Prefijo configurable (`mqtt.topic_prefix`):

| Tópico | QoS | Retained | Descripción |
|--------|-----|----------|-------------|
| `<prefix>/telemetry` | 1 | false | Snapshot por polling |
| `$CONTROL/health`     | 1 | true  | LWT; payload `online`/`offline` |
| `$CONTROL/inventory`  | 1 | true  | Info del device (model, fw, etc.) |

## Payload (schema v0.1.0)

```json
{
  "schema_version": "0.1.0",
  "snapshot": {
    "ts": "2026-09-07T12:00:00Z",
    "model": "sunspec",
    "dev": "sunspec-192.168.1.50",
    "status": "ok",
    "metrics": {
      "ac.w": 0,
      "ac.var": 0,
      "ac.va": 0,
      "freq": 60,
      "dc.w": 0,
      "dc.v": 0,
      "dc.a": 0
    }
  }
}
```

## Versionado
- Campo `schema_version` siempre presente.
- Cambios incompatibles ⇒ bump mayor (`1.0.0`).

## Seguridad
- mTLS con `ca_file`, `client_cert`, `client_key`.
- `insecure_skip_verify` solo en dev, con warning en logs.
