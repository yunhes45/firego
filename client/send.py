import websocket

SESSION_ID = "AB3ADC7BAF5DDB0F1E7E17232B1A0A84"
FILE_PATH = "testfile.txt"

def send_file():
    ws = websocket.WebSocket()
    ws.connect(f"ws://localhost:54321/send/{SESSION_ID}")
    print("연결됨, READY 대기중...")

    msg = ws.recv()
    print(f"서버 신호: {msg}")

    if msg == "READY":
        print("전송 시작!")
        with open(FILE_PATH, "rb") as f:
            data = f.read()
            ws.send_binary(data)
            print(f"전송완료 ({len(data)} bytes)")
        ws.close()

send_file()