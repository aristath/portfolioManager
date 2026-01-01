FROM python:3.13-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*

# Copy requirements and install dependencies
COPY services/evaluator/requirements.txt /app/services/evaluator/
RUN pip install --no-cache-dir -r /app/services/evaluator/requirements.txt

# Copy application code
COPY app /app/app/
COPY services/evaluator /app/services/evaluator/

# Set environment
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Expose port (configurable via EVALUATOR_PORT env var)
EXPOSE 8010

# Run service
CMD ["python", "-m", "services.evaluator.main"]
