FROM alpine:3.13

# Install gcc and mysql
RUN apk add gcc

# Create login_server directory
RUN mkdir ship_server
COPY ./ ./ship_server
WORKDIR ship_server

# WHAT CONFIG FILE(S) TO COPY?

# Compile the code
RUN gcc ship_server.c -o ship_server

# Copy login server executable for the server files
ENTRYPOINT ["./ship_server"]

# TODO:
# compile account_add
# compile make_key
