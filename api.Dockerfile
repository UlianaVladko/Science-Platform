FROM python:3.12-slim

WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY vendor/literature-api/pyproject.toml ./vendor/literature-api/pyproject.toml
COPY vendor/literature-api/src ./vendor/literature-api/src

RUN pip install --no-cache-dir -e ./vendor/literature-api

EXPOSE 8000
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000", "--app-dir", "vendor/literature-api/src"]
