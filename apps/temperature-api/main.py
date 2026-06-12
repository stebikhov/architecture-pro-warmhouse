import logging
import os
import random
import sys
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException, Query

logging.basicConfig(
    level=logging.INFO,
    format="[%(levelname)s] %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger(__name__)

SENSOR_TO_LOCATION = {
    "1": "Living Room",
    "2": "Bedroom",
    "3": "Kitchen",
    "4": "Bathroom",
    "5": "Garage",
    "6": "Office",
    "7": "Dining Room",
    "8": "Basement",
    "9": "Attic",
    "10": "Guest Room",
}

LOCATION_TO_SENSOR = {v: k for k, v in SENSOR_TO_LOCATION.items()}

app = FastAPI(title="Temperature API")


@app.get("/health")
def health_check():
    logger.info("[REQUEST] Health check requested")
    return {"status": "ok"}


@app.get("/temperature")
def get_temperature(
    location: str = Query(default=""),
    sensorId: str = Query(default=""),
):
    logger.info(
        "[HANDLER] getTemperature called - location: '%s', sensorId: '%s'",
        location,
        sensorId,
    )

    if location == "":
        location = SENSOR_TO_LOCATION.get(sensorId, "Unknown")
        logger.info("[LOGIC] Mapped sensorId '%s' to location '%s'", sensorId, location)

    if sensorId == "":
        sensorId = LOCATION_TO_SENSOR.get(location, "0")
        logger.info("[LOGIC] Mapped location '%s' to sensorId '%s'", location, sensorId)

    temp = round(15 + random.random() * 15, 2)
    logger.info("[LOGIC] Generated random temperature: %.2f", temp)

    return {
        "sensor_id": sensorId,
        "location": location,
        "value": temp,
        "unit": "\u00b0C",
        "status": "active",
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }


@app.get("/temperature/{sensor_id}")
def get_temperature_by_id(sensor_id: str):
    logger.info("[HANDLER] getTemperatureByID called - id: '%s'", sensor_id)

    location = SENSOR_TO_LOCATION.get(sensor_id)
    if location is None:
        logger.info("[ERROR] Sensor '%s' not found - returning 404", sensor_id)
        raise HTTPException(status_code=404, detail="sensor not found")

    temp = round(15 + random.random() * 15, 2)
    logger.info("[LOGIC] Generated random temperature: %.2f", temp)

    return {
        "sensor_id": sensor_id,
        "location": location,
        "value": temp,
        "unit": "\u00b0C",
        "status": "active",
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }
