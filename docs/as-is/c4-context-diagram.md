# Smart Home System - Context Diagram (C4 Level 1)

## Overview

This diagram shows the Smart Home System in its current monolithic state, depicting the interactions between users, the system, and external dependencies.

## Actors

- **Пользователь** — Владелец умного дома, управляющий устройствами и просматривающий данные датчиков
- **Администратор** — Технический специалист, настраивающий систему и управляющий устройствами

## System

- **Smart Home Monolith** — Go + Gin + PostgreSQL, управление датчиками, сбор телеметрии, интеграция с внешними API

## External Dependencies

- **Temperature API** — Внешний сервис, предоставляющий данные о температуре в реальном времени
- **PostgreSQL** — База данных для хранения данных о датчиках и их показаниях

## Interactions

| From | To | Description | Protocol |
|------|-----|-------------|----------|
| Пользователь | Smart Home Monolith | Использует | HTTP/REST |
| Администратор | Smart Home Monolith | Настраивает | HTTP/REST |
| Smart Home Monolith | Temperature API | Запрашивает данные о температуре | HTTP/REST |
| Smart Home Monolith | PostgreSQL | Чтение/запись данных | PostgreSQL Protocol |
