# docker-volume-nas

Simple NAS volume manager:
Create volume on shared file system based on folder.

Based on the simple practice to mount a shared volume (i.e. via nfs) and do bind volumes on different repertories in it.

The plugin will look for a mount point and create a new folder in it for each new volume.

## Working with nfs share and bind

A common way to persist data for containers is to use bind and nfs/cifs/fuse mount point

```ascii
  +---------------+ +---------------+
  | docker host   | | docker host   |
  | +-----------+ | | +-----------+ |
  | | container | | | | container | |
  | | /foo      | | | | /foo      | |
  | +----^------+ | | +----^------+ |
  |      | bind   | |      | bind   |
  | /var/nfs/foo  | | /var/nfs/foo  |
  +------^--------+ +------^--------+
         | nfs/cifs mount  |
         +---------+-------+
                   |
      +------------|----+
      | nas server |    |
      | exports /docker |
      +-----------------+
```

By using a volume plugin the folders can be automaticaly created.

```ascii
  +---------------+ +---------------+
  | docker host   | | docker host   |
  | +-----------+ | | +-----------+ |
  | | container | | | | container | |
  | | /foo      | | | | /foo      | |
  | +----^------+ | | +----^------+ |
  |      | volume | |      | volume |
  |      |  "foo" | |      |  "foo" |
  | /var/nfs/foo  | | /var/nfs/foo  |
  +------^--------+ +------^--------+
         | nfs/cifs mount  |
         +---------+-------+
                   |
      +------------|----+
      | nas server |    |
      | exports /docker |
      +-----------------+
```

# TODO / Investigation

An evolution of that would be to mount specific folders only when a container needs it. A bit like automapping does. This helps reduce the number of mounts open on an nfs export. Additionally a host not running a container won't have the nfs share mounted.
Tell me if intresting, otherwise i will take my time on it.
