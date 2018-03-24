FROM scratch
ADD lxcfs-initializer /lxcfs-initializer
ENTRYPOINT ["/lxcfs-initializer"]
