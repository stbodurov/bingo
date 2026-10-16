FROM python:3.12-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 HARBORMASTER_DATA=/data HARBORMASTER_DB=sqlite:////data/harbormaster.db
WORKDIR /srv
COPY pyproject.toml ./
COPY harbormaster ./harbormaster
RUN pip install --no-cache-dir . && useradd -r -u 10001 hm && mkdir -p /data && chown hm /data
USER hm
VOLUME /data
EXPOSE 8000
HEALTHCHECK CMD python -c "import urllib.request;urllib.request.urlopen('http://127.0.0.1:8000/healthz')"
CMD ["uvicorn", "harbormaster.main:app", "--host", "0.0.0.0", "--port", "8000", "--proxy-headers", "--workers", "2"]
