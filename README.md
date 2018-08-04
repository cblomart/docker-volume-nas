# docker-volume-nas

Simple NAS volume manager:
Create volume on shared file system based on folder.

Based on the simple practice to mount a shared volume (i.e. via nfs) and do bind volumes on different repertories in it.

The plugin will look for a mount point and create a new folder in it for each new volume.

This plugin is made to avoid the necessity to create a new share or folder in a share each time a permanent shared storage is needed.
