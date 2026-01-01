FROM python:3.13-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*

# Copy requirements and install dependencies
COPY services/opportunity/requirements.txt /app/services/opportunity/
RUN pip install --no-cache-dir -r /app/services/opportunity/requirements.txt

# Copy application code
COPY app /app/app/
COPY services/opportunity /app/services/opportunity/

# Set environment
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# Expose port
EXPOSE 8008

# Run service
CMD ["python", "-m", "services.opportunity.main"]
