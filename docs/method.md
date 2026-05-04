## connect

```json
{
  "type": "req",
  "id": "req-1-1777890202952",
  "method": "connect",
  "params": {
    "token": "9c78a321c2279fabeba00abd1b5b6fc2",
    "user_id": "system",
    "sender_id": "",
    "locale": "en",
    "tenant_hint": "",
    "tenant_id": "",
    "protocolVersion": 3
  }
}
```

```json
{
  "type": "res",
  "id": "req-1-1777890202952",
  "ok": true,
  "payload": {
    "protocol": 3,
    "role": "user",
    "server": {
      "name": "pomclaw",
      "version": "0.1.0"
    },
    "user_id": "system"
  }
}
```

#### goclaw 正确返回

```json
{
  "type": "res",
  "id": "req-1-1777890922788",
  "ok": true,
  "payload": {
    "edition": "lite",
    "is_master_scope": true,
    "is_owner": true,
    "protocol": 3,
    "role": "owner",
    "server": {
      "name": "goclaw",
      "version": "dev"
    },
    "tenant_id": "0193a5b0-7000-7000-8000-000000000001",
    "user_id": "system"
  }
}
```

