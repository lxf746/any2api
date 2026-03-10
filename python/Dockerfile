FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    NEWPLATFORM2API_HOST=0.0.0.0 \
    NEWPLATFORM2API_PORT=8100 \
    NEWPLATFORM2API_ADMIN_STORE_PATH=/app/data/admin.json

COPY admin_store.py gateway_types.py pyproject.toml server.py ./
COPY providers ./providers

EXPOSE 8100

CMD ["python", "server.py"]