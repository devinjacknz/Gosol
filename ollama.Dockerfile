FROM ollama/ollama:latest

# Install curl and other necessary tools
RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*

# Copy model initialization script
COPY <<EOF /init.sh
#!/bin/sh
ollama serve &
sleep 10
ollama pull deepseek-r1:1.5b
exec ollama serve
EOF

RUN chmod +x /init.sh
ENTRYPOINT ["/init.sh"]
