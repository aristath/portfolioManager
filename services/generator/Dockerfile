FROM python:3.13-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*

# Copy requirements and install dependencies
COPY services/generator/requirements.txt /app/services/generator/
RUN pip install --no-cache-dir -r /app/services/generator/requirements.txt

# Copy application code
COPY app /app/app/
COPY services/generator /app/services/generator/

# Set environment
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Expose port
EXPOSE 8009

# Run service
CMD ["python", "-m", "services.generator.main"]
