# SunSpec — Modelos soportados (Sprint 0)

## Implementación
- Transporte: Modbus TCP (puerto 502).
- Validación del magic `SunS` en registros base (`40000-40001`).
- Lectura por polling configurable (`poll.interval`).
- CRC16 Modbus en cada PDU.

## Modelos (mapping a `Snapshot.Metrics`)

| Modelo | Métricas |
|--------|----------|
| 101 (AC)            | `ac.w`, `ac.var`, `ac.va`, `freq` |
| 103 (Status)        | `status` |
| 120 (DC strings)    | `dc.w`, `dc.v`, `dc.a` |
| 121 (DC single)     | `dc.w`, `dc.v`, `dc.a` |
| 123 (DC advanced)   | `dc.w`, `dc.v`, `dc.a` |

> Sprint 0 incluye el handshake y stub de mapping. La decodificación bit-a-bit de cada modelo se completará en Sprint 1 (SunspecDecoder real).

## Añadir un modelo nuevo (preparación para drivers propietarios)
1. Crear un nuevo paquete en `internal/inverter/<vendor>/`.
2. Implementar `inverter.Driver` (`Scan`, `Model`, `DeviceID`, `Close`).
3. Registrar en `internal/inverter/registry.go` bajo un nuevo `type` en `config.example.yaml`.
4. Añadir tests unitarios y (si aplica) un simulador en `scripts/`.
