import hmac, hashlib, base64, time, uuid

SECRET = "key-@@@@)))()((9))-xxxx&&&%%%%%"

def generate_signature(message, timestamp=None, request_id=None, user_id="anonymous"):
    if timestamp is None:
        timestamp = str(int(time.time() * 1000))
    if request_id is None:
        request_id = str(uuid.uuid4())

    # 1. sortedPayload
    entries = sorted({"timestamp": timestamp, "requestId": request_id, "user_id": user_id}.items())
    sorted_payload = ",".join(f"{k},{v}" for k, v in entries)

    # 2. base64Prompt
    base64_prompt = base64.b64encode(message.encode("utf-8")).decode("ascii")

    # 3. 待签名字符串
    d = f"{sorted_payload}|{base64_prompt}|{timestamp}"

    # 4. 时间窗口
    w = int(timestamp) // 300000

    # 5. 动态密钥
    E = hmac.new(SECRET.encode(), str(w).encode(), hashlib.sha256).hexdigest()

    # 6. 最终签名
    return hmac.new(E.encode(), d.encode(), hashlib.sha256).hexdigest()

# 测试
ts = str(int(time.time() * 1000))
rid = str(uuid.uuid4())
sig = generate_signature("你好", ts, rid)
print(f"timestamp:  {ts}")
print(f"requestId:  {rid}")
print(f"signature:  {sig}")
print(f"长度: {len(sig)} (应该是 64)")
