import websocket

SESSION_ID = "CBD993"
OUTPUT_PATH = "received_file.txt"

def receive_file():
    ws = websocket.WebSocket()
    ws.connect(f"ws://localhost:54321/receive/{SESSION_ID}")
    print("연결됨, 파일 대기중...")

    with open(OUTPUT_PATH, "wb") as f:
        while True:
            try:
                data = ws.recv()
                if not data:
                    break
                # 앞 8바이트 청크 번호 제거하고 저장
                f.write(data[8:] if isinstance(data, bytes) else data.encode())
                print(f"{len(data)-8} bytes 받음")
            except Exception as e:
                print(f"수신 완료: {e}")
                break

    print(f"파일 저장완료: {OUTPUT_PATH}")

receive_file()