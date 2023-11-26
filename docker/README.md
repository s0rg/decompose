# decompose

Reverse-engineering tool for docker environments.

# how to run

```
docker run \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v /:/rootfs:ro \
        -e IN_DOCKER_PROC_ROOT=/rootfs \
        s0rg/decompose:latest -format stat
```

[more options and documentaion](https://github.com/s0rg/decompose)
