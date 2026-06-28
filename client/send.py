import websocket
import threading

GROUP_ID = "EAE43D2C136CEA34F65D1273255493C0"
FILES = [
    {"file_id": "D1710616F694174C57AE2EAB91E31EC4", "path": "a.zip"},
    {"file_id": "19607C22424DF6FA0B5C438C8AD8E903", "path": "b.zip"},
]

def send_file(group_id, file_id, file_path):
    ws = websocket.WebSocket()
    ws.connect(f"ws://localhost:54321/send/{group_id}/{file_id}")
    print(f"{file_path} 연결됨, READY 대기중...")

    msg = ws.recv()
    if msg == "READY":
        print(f"{file_path} 전송 시작!")
        with open(file_path, "rb") as f:
            data = f.read()
            ws.send_binary(data)
            print(f"{file_path} 전송완료 ({len(data)} bytes)")
    ws.close()

threads = []
for f in FILES:
    t = threading.Thread(target=send_file, args=(GROUP_ID, f["file_id"], f["path"]))
    threads.append(t)
    t.start()

for t in threads:
    t.join()