FROM registry.cern.ch/docker.io/library/fedora:38

RUN dnf install -y fuse3 fuse3-devel \
    && mkdir -p /lib64/autofs \
    && ln -sf /usr/local/bin/dummy-fuse /sbin/mount.dummy-fuse \
    && ln -sf /lib64/autofs/lookup_file.so /lib64/autofs/lookup_files.so \
    && ln -sf /usr/local/bin/dummy-fuse /sbin/mount.dummy-fuse

# For debugging...
RUN dnf -y install gdb git procps

COPY dist/ /

COPY workload/build/dummy-fuse-workload /usr/local/bin

COPY csi/build/automount-reconciler /usr/local/bin
COPY csi/build/automount-runner /usr/local/bin
COPY csi/build/csi-driver /usr/local/bin

COPY autofs/lib/libautofs.so /lib64
COPY autofs/modules/*.so /lib64/autofs
COPY autofs/daemon/automount /usr/local/bin

COPY fs/build/dummy-fuse /usr/local/bin
