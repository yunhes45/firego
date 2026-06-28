import websocket
import threading
import time

GROUP_ID = "EAE43D2C136CEA34F65D1273255493C0"
FILES = [
    {"file_id": "D1710616F694174C57AE2EAB91E31EC4", "output": "received_a.zip"},
    {"file_id": "19607C22424DF6FA0B5C438C8AD8E903", "output": "received_b.zip"},
]

def receive_file(group_id, file_id, output_path):
    ws = websocket.WebSocket()
    ws.connect(f"ws://localhost:54321/receive/{group_id}/{file_id}")
    print(f"{output_path} 대기중...")

    with open(output_path, "wb") as f:
        while True:
            try:
                data = ws.recv()
                if not data:
                    break
                f.write(data[8:] if isinstance(data, bytes) else data.encode())
                print(f"[{time.strftime('%H:%M:%S')}] {output_path} {len(data)-8} bytes 받음")
            except Exception as e:
                print(f"{output_path} 수신 완료: {e}")
                break

    print(f"[{time.strftime('%H:%M:%S')}] {output_path} 저장완료")

threads = []
for f in FILES:
    t = threading.Thread(target=receive_file, args=(GROUP_ID, f["file_id"], f["output"]))
    threads.append(t)
    t.start()

for t in threads:
    t.join()