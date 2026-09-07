import socket
import struct
import time
import threading

MAGIC = b"SunS"

def modbus_crc(data: bytes) -> bytes:
    crc = 0xFFFF
    for b in data:
        crc ^= b
        for _ in range(8):
            if crc & 1:
                crc = (crc >> 1) ^ 0xA001
            else:
                crc >>= 1
    return struct.pack("<H", crc)

def build_response(pdu: bytes, unit_id: int = 1) -> bytes:
    mbap = struct.pack(">HHHB", 0, len(pdu) + 2, 0, unit_id)
    return mbap + pdu + modbus_crc(pdu)

def handle(conn, addr):
    print(f"simulator: client {addr}")
    try:
        while True:
            header = conn.recv(8)
            if not header:
                break
            tx_id, proto, length, unit_id = struct.unpack(">HHHB", header)
            pdu = b""
            while len(pdu) < length - 2:
                chunk = conn.recv(length - 2 - len(pdu))
                if not chunk:
                    break
                pdu += chunk
            if len(pdu) < 1:
                continue
            func = pdu[0]
            if func == 3:
                reg = struct.unpack(">H", pdu[1:3])[0]
                qty = struct.unpack(">H", pdu[3:5])[0]
                if reg == 40000 and qty >= 2:
                    payload = MAGIC + struct.pack(">HH", 1, 65)
                    pdu_out = bytes([3, len(payload)]) + payload + b"\x00" * 60
                else:
                    values = []
                    for i in range(qty):
                        values.append(struct.pack(">H", 1000 + i))
                    payload = b"".join(values)
                    pdu_out = bytes([3, len(payload)]) + payload
                conn.sendall(build_response(pdu_out, unit_id))
            elif func == 6:
                pdu_out = pdu
                conn.sendall(build_response(pdu_out, unit_id))
            else:
                err = bytes([func | 0x80, 1])
                conn.sendall(build_response(err, unit_id))
    except Exception as e:
        print(f"simulator: {e}")
    finally:
        conn.close()

def main():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", 5020))
    s.listen(5)
    print("simulator: listening on 0.0.0.0:5020")
    while True:
        c, a = s.accept()
        threading.Thread(target=handle, args=(c, a), daemon=True).start()

if __name__ == "__main__":
    main()
