FROM alpine:3.13

# Install gcc and mysql
RUN apk add gcc

# Create login_server directory
RUN mkdir login_server
COPY ./ ./login_server
WORKDIR login_server

# WHAT CONFIG FILE(S) TO COPY?

# Compile the code
RUN gcc login_server.c -o login_server

# Copy login server executable for the server files
ENTRYPOINT ["./login_server"]

# TODO:
# compile account_add
# compile make_key
