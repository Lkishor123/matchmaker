import os
import json
import base64
import requests
import subprocess
import tempfile
import websocket
import pytest

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")
AUTH_URL = os.getenv("AUTH_SERVICE_URL", "http://localhost:8081")
CHAT_URL = os.getenv("CHAT_SERVICE_URL", "http://localhost:8082")
MATCH_URL = os.getenv("MATCH_SERVICE_URL", "http://localhost:8083")
USER_URL = os.getenv("USER_SERVICE_URL", "http://localhost:8084")
REPORT_URL = os.getenv("REPORT_SERVICE_URL", "http://localhost:8085")


def _b64(data: bytes) -> bytes:
    return base64.urlsafe_b64encode(data).rstrip(b"=")


def dummy_token(user_id: int = 1) -> str:
    header = _b64(json.dumps({"alg": "none", "typ": "JWT"}).encode())
    payload = _b64(json.dumps({"user_id": user_id}).encode())
    return (header + b"." + payload + b".").decode()


def signed_token(private_key: str, user_id: int = 1) -> str:
    header = _b64(json.dumps({"alg": "RS256", "typ": "JWT"}).encode())
    payload = _b64(json.dumps({"user_id": user_id}).encode())
    message = header + b"." + payload
    with tempfile.NamedTemporaryFile("wb", delete=False) as f:
        f.write(private_key.encode())
        path = f.name
    try:
        proc = subprocess.run(
            ["openssl", "dgst", "-sha256", "-sign", path],
            input=message,
            stdout=subprocess.PIPE,
            check=True,
        )
    finally:
        os.unlink(path)
    signature = _b64(proc.stdout)
    return (message + b"." + signature).decode()


def ws_url(base: str, path: str) -> str:
    return base.replace("http", "ws", 1) + path


@pytest.mark.parametrize(
    "base",
    [GATEWAY_URL, AUTH_URL, CHAT_URL, MATCH_URL, USER_URL, REPORT_URL],
)
def test_ping(base):
    r = requests.get(base + "/ping")
    assert r.status_code == 200
    assert r.json() == {"message": "pong"}


def test_user_service():
    payload = {"email": "pytest@example.com", "name": "Tester"}
    r = requests.post(USER_URL + "/internal/v1/users", json=payload)
    assert r.status_code in (200, 201)
    assert "id" in r.json()

    r = requests.post(USER_URL + "/internal/v1/users", json={"name": "bad"})
    assert r.status_code == 400


def test_report_service():
    good = {"dob": "2000-01-01", "tob": "12:00:00", "lat": 1, "lon": 2}
    r = requests.post(REPORT_URL + "/internal/v1/reports", json=good)
    assert r.status_code == 200

    bad = {"dob": "nope"}
    r = requests.post(REPORT_URL + "/internal/v1/reports", json=bad)
    assert r.status_code == 400


def test_match_service():
    payload = {
        "personA": {"dob": "2000-01-01", "tob": "12:00:00", "lat": 1, "lon": 2},
        "personB": {"dob": "2001-01-01", "tob": "11:00:00", "lat": 1, "lon": 2},
    }
    r = requests.post(MATCH_URL + "/api/v1/analysis", json=payload)
    assert r.status_code == 200
    assert "score" in r.json()

    r = requests.post(MATCH_URL + "/api/v1/analysis", json={"foo": "bar"})
    assert r.status_code == 400


def test_chat_service():
    token = dummy_token()
    ws = websocket.create_connection(
        ws_url(CHAT_URL, "/api/v1/chat"),
        header={"Authorization": f"Bearer {token}"},
    )
    ws.send("hello")
    msg = ws.recv()
    assert msg
    ws.close()


def test_chat_service_llm_failure():
    token = dummy_token()
    ws = websocket.create_connection(
        ws_url(CHAT_URL, "/api/v1/chat"),
        header={"Authorization": f"Bearer {token}"},
    )
    ws.send("trigger-error")
    with pytest.raises(Exception):
        ws.recv()
    ws.close()


def test_gateway_proxy():
    key = os.getenv("JWT_PRIVATE_KEY")
    if not key:
        pytest.skip("JWT_PRIVATE_KEY not configured")
    token = signed_token(key, 1)
    headers = {"Authorization": f"Bearer {token}"}

    r = requests.get(GATEWAY_URL + "/api/v1/users/me", headers=headers)
    assert r.status_code in (200, 404)

    payload = {
        "personA": {"dob": "2000-01-01", "tob": "12:00:00", "lat": 1, "lon": 2},
        "personB": {"dob": "2001-01-01", "tob": "11:00:00", "lat": 1, "lon": 2},
    }
    r = requests.post(GATEWAY_URL + "/api/v1/analysis", json=payload, headers=headers)
    assert r.status_code in (200, 400)

    ws = websocket.create_connection(
        ws_url(GATEWAY_URL, "/api/v1/chat"),
        header={"Authorization": f"Bearer {token}"},
    )
    ws.send("hi")
    ws.recv()
    ws.close()


def test_gateway_invalid_token():
    headers = {"Authorization": "Bearer invalid"}
    r = requests.get(GATEWAY_URL + "/api/v1/users/me", headers=headers)
    assert r.status_code == 401

    r = requests.post(GATEWAY_URL + "/api/v1/analysis", json={}, headers=headers)
    assert r.status_code == 401

    with pytest.raises(websocket.WebSocketBadStatusException):
        websocket.create_connection(ws_url(GATEWAY_URL, "/api/v1/chat"), header=headers)


