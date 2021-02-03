FROM alpine:3.13

# Install gcc and mysql
RUN apk add gcc

# Create login_server directory
RUN mkdir patch_server
COPY ./ ./patch_server
WORKDIR patch_server

# WHAT CONFIG FILE(S) TO COPY?

# Compile the code
RUN gcc patch_server.c -o patch_server

# Copy login server executable for the server files
ENTRYPOINT ["./patch_server"]

# TODO:
# compile account_add
# compile make_key
