# Stage 0
# FROM debian AS lib_builder

# WORKDIR /foundry

# RUN apt-get update -y
# RUN apt-get install -y \
#     portaudio19-dev \
#     gccgo-go \
#     pkg-config \
#   build-essential \
#   cmake \
    # git

# RUN git clone https://github.com/jgarff/rpi_ws281x.git \
#   && cd rpi_ws281x \ 
#   && mkdir build \
#   && cd build \ 
#   && cmake -D BUILD_SHARED=OFF -D BUILD_TEST=OFF .. \
#   && cmake --build . \
#   && make install

# Stage 1 : Build a go image with the rpi_ws281x C library and the go wrapper

FROM golang:1.17.9
RUN apt-get update -y
RUN apt-get install -y \
    portaudio19-dev \
    libmpg123-dev \
    libmp3lame-dev \
    gccgo-go \
    pkg-config \
    build-essential \
    cmake \
    git
# COPY --from=lib_builder /usr/local/lib/libws2811.a /usr/local/lib/
# COPY --from=lib_builder /usr/local/include/ws2811 /usr/local/include/ws2811

# RUN go get github.com/rpi-ws281x/rpi-ws281x-go