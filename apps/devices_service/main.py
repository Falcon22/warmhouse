from datetime import datetime, timezone
from typing import Dict, Optional
from uuid import uuid4

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel


class DeviceIn(BaseModel):
    name: str
    type: str
    location: str


class Device(DeviceIn):
    id: str
    status: str
    created_at: datetime


class DeviceRegistry:
    def __init__(self) -> None:
        self._devices: Dict[str, Device] = {}

    def list(self) -> list[Device]:
        return list(self._devices.values())

    def get(self, device_id: str) -> Optional[Device]:
        return self._devices.get(device_id)

    def create(self, payload: DeviceIn) -> Device:
        device = Device(
            id=str(uuid4()),
            name=payload.name,
            type=payload.type,
            location=payload.location,
            status="active",
            created_at=datetime.now(timezone.utc),
        )
        self._devices[device.id] = device
        return device

    def delete(self, device_id: str) -> bool:
        return self._devices.pop(device_id, None) is not None


app = FastAPI(title="Devices Service")
registry = DeviceRegistry()


@app.get("/health")
def health() -> dict:
    return {"status": "ok"}


@app.get("/devices")
def list_devices() -> list[Device]:
    return registry.list()


@app.get("/devices/{device_id}")
def get_device(device_id: str) -> Device:
    device = registry.get(device_id)
    if device is None:
        raise HTTPException(status_code=404, detail="Device not found")
    return device


@app.post("/devices", status_code=201)
def create_device(payload: DeviceIn) -> Device:
    return registry.create(payload)


@app.delete("/devices/{device_id}", status_code=204)
def delete_device(device_id: str) -> None:
    if not registry.delete(device_id):
        raise HTTPException(status_code=404, detail="Device not found")
