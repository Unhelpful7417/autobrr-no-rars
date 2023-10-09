# Pin base image to version to avoid stability issues
FROM python:3.11-slim-bookworm

# Install system dependencies
RUN apt update && apt install -y python3-libtorrent && rm -rf /var/lib/apt/lists/*

# Set variables
ENV DEBIAN_FRONTEND=noninteractive
## Helps python find libtorrent so we can import it
ENV PYTHONPATH=/usr/lib/python3/dist-packages:$PYTHONPATH
ENV serverPort=8080
ENV logLevel=info

# Set working directory
WORKDIR /app

# Copy in python dependencies
COPY requirements.txt ./

# Install python dependencies
RUN pip install -r requirements.txt

# Copy in main script
COPY api.py ./

# Run the API server
CMD ["bash", "-c", "uvicorn api:app --host 0.0.0.0 --log-level ${logLevel} --port ${serverPort}"]
EXPOSE ${serverPort}