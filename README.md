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

By using a volume plugin the folders can be automaticaly created and removed.

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