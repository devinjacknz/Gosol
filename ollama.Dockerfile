FROM ollama/ollama:latest

# Install curl and other necessary tools
RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*
