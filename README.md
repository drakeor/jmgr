# JMGR - Jettison Server Manager

JMGR, or the jettison server manager, is a command line server management tool for grabbing the status of multiple servers at once.

We built this server to fulfill the following requirements:

* Mostly graphical interface, kind of like running top. Mostly built for Linux.

* Runs in a closed SSH-only environment, with no access to a web interface

* Can talk to other servers in a whitelist, allows us to get the status of multiple servers at once without ssh-ing into each one

* Plug-in modules to get status of things like QEMU virtual machines



# Roadmap

* Get a service running, client connects and gets ping and says if service is live

* Get current CPU, ram, network, disk usage for servers

* Get current virtual machines on the servers

* Get processes on each server

* Services store historical data

* Services keep historical data for some set amount of time

# Credits

https://github.com/stefanhaustein/TerminalImageViewer

For terminal image viewer, which helps display images in the console

https://github.com/ljrrjl/ftxui-image-view

For simple terminal image viewer integration to FTXUI

# Pre-requisites

ImageMagick: (For image displays):

```bash
apt-get install imagemagick
```

ZeroMQ: 

```bash
apt-get install libzmq3-dev
```

Since we use gRPC, you need the following:

```bash
 $ [sudo] apt-get install build-essential autoconf libtool pkg-config
```

And of course, cmake

```bash
[sudo] apt-get install cmake
```

# Building the project

Sometimes gRPC doesn't play super nicely. In that case, we need to install it and link it with a different CMake somehow.

```bash
cmake .
make
bin/jmgr
```

# Usage

There's going to be a whitelist file (or add a link to a whitelist file on a shared device or something idc)

The server ips should be added to this whitelist file. Can also set it to accept from a range if you want.

Then you just run the server manager. 

Press q to quit the server manager.

You can also set bash to start the server manager whenever you log in. And then press 'q' to exit. Or Control-C.