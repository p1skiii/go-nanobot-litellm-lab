FROM scratch

WORKDIR /app
COPY deploy/bin/nanobot /app/nanobot
COPY configs /app/configs

EXPOSE 8080
ENTRYPOINT ["/app/nanobot"]
