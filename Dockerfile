FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y \
    cmake \
    git \
    build-essential \
    libtool \
    pkg-config \
    autoconf \
    automake \
    libssl-dev \
    g++ \ 
    libzmq3-dev \
    imagemagick

# Install ZeroMQ from source
RUN git clone https://github.com/zeromq/libzmq.git && \
    cd libzmq && \
    mkdir build && \
    cd build && \
    cmake .. && \
    make -j$(nproc) && \
    make install && \
    ldconfig

# Copy the source code
COPY . /app

# Build the application
WORKDIR /app
RUN cmake .
RUN make -j$(nproc)

# Run the splash application
CMD ["bin/jmgr-splash"]
